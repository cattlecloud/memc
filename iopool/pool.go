// Copyright (c) CattleCloud LLC
// SPDX-License-Identifier: BSD-3-Clause

package iopool

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"cattlecloud.net/go/scope"
	"cattlecloud.net/go/stacks"
)

var (
	ErrClientClosed = errors.New("memc: client has been closed")
)

// A Connection represents an underlying TCP/Unix socket connection to a single
// memcached instance.
//
// It may be reused in future requests.
type Connection interface {
	// Read reads data from the connection.
	Read(b []byte) (n int, err error)

	// Write writes data to the connection.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	Close() error
}

type Buffer struct {
	*bufio.Reader
	*bufio.Writer
	io.Closer
	failure *atomic.Bool
}

func newBuffer(conn Connection) *Buffer {
	return &Buffer{
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
		conn,
		new(atomic.Bool),
	}
}

func (b *Buffer) SetHealth(err error) {
	if err != nil {
		b.failure.Store(true)
	}
}

func New(instances []string, idle int) *Collection {
	pools := make([]*pool, 0, len(instances))
	for _, instance := range instances {
		pools = append(pools, newPool(instance, idle))
	}
	return &Collection{pools: pools}
}

type Collection struct {
	pools []*pool
}

func (c *Collection) pick(key string) int {
	if len(c.pools) == 1 {
		return 0
	}

	// compute the server to choose for key
	// deterministic given set of servers and key
	x := byte(37)
	for _, c := range key {
		x ^= byte(c)
	}
	idx := int(int(x) % len(c.pools))

	return idx
}

func (c *Collection) Get(key string) (*Buffer, error) {
	idx := c.pick(key)
	choice := c.pools[idx]
	return choice.get()
}

func (c *Collection) Return(key string, conn *Buffer) {
	idx := c.pick(key)
	choice := c.pools[idx]
	choice.free(conn)
}

func (c *Collection) Close() error {
	for _, p := range c.pools {
		p.close()
	}
	return nil
}

const closed = -1

type pool struct {
	address   string
	available stacks.Stack[*Buffer]
	idle      int
	openf     func(string) (Connection, error)
}

func newPool(address string, idle int) *pool {
	return &pool{
		address:   address,
		idle:      idle,
		openf:     open,
		available: stacks.Simple[*Buffer](),
	}
}

func (p *pool) close() {
	p.idle = closed // close down the pool

	// pop off each idle connection and close it
	for !p.available.Empty() {
		conn := p.available.Pop()
		_ = conn.Close()
	}
}

func (p *pool) get() (*Buffer, error) {
	if p.idle == closed {
		return nil, ErrClientClosed
	}

	if p.available.Empty() {
		conn, err := p.openf(p.address)
		if err != nil {
			return nil, err
		}
		return newBuffer(conn), nil
	}

	b := p.available.Pop()
	return b, nil
}

func open(address string) (Connection, error) {
	dialer := &net.Dialer{Timeout: 3 * time.Second}

	ctx, cancel := scope.TTL(3 * time.Second)
	defer cancel()

	switch strings.HasPrefix(address, "/") {
	case true:
		return dialer.DialContext(ctx, "unix", address)
	default:
		return dialer.DialContext(ctx, "tcp", address)
	}
}

func (p *pool) free(conn *Buffer) {
	switch {
	case p.idle == closed:
		_ = conn.Close()
	case p.available.Size() >= p.idle:
		_ = conn.Close()
	case conn.failure.Load():
		_ = conn.Close()
	default:
		p.available.Push(conn)
	}
}
