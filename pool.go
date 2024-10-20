// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"net"
	"sort"
	"time"

	"noxide.lol/go/stacks"
)

type pools struct {
	servers []*pool
}

func (p *pools) close() {
	for _, server := range p.servers {
		server.close()
	}
}

func (p *pools) create(addresses []string, timeout time.Duration, idle int) {
	sort.Strings(addresses) // helps with determinism
	for _, address := range addresses {
		p.add(address, timeout, idle)
	}
}

func (p *pools) add(address string, timeout time.Duration, idle int) {
	p.servers = append(p.servers, &pool{
		address:   address,
		timeout:   timeout,
		idle:      idle,
		available: stacks.Simple[net.Conn](),
	})
}

func (p *pools) pick(key string) int {
	if len(p.servers) == 1 {
		return 0
	}

	// compute the server to choose for key
	// deterministic given set of servers and key
	x := byte(37)
	for _, c := range key {
		x ^= byte(c)
	}
	idx := int(int(x) % len(p.servers))

	return idx
}

type donef func()

func (p *pools) get(key string) (net.Conn, donef, error) {
	idx := p.pick(key)
	spool := p.servers[idx]
	return spool.get()
}

type pool struct {
	timeout   time.Duration
	address   string
	available stacks.Stack[net.Conn]
	idle      int
}

// closed is a sentinel value that no connections should be kept as
// idle connections and simply closed once they're done
const closed = -1

func (p *pool) close() {
	p.idle = closed // close down the pool

	// pop off each idle connection and close it
	for !p.available.Empty() {
		conn := p.available.Pop()
		_ = conn.Close()
	}
}

func (p *pool) get() (net.Conn, donef, error) {
	if p.idle == closed {
		return nil, nil, ErrClientClosed
	}

	if p.available.Empty() {
		conn, err := p.open()
		if err != nil {
			return nil, nil, err
		}
		done := func() { p.discard(conn) }
		return conn, done, nil
	}

	conn := p.available.Pop()
	done := func() { p.discard(conn) }
	return conn, done, nil
}

func (p *pool) open() (net.Conn, error) {
	return net.DialTimeout("tcp", p.address, p.timeout)
}

func (p *pool) discard(conn net.Conn) {
	// if the pool is closed or there are already enough idle
	// connections, just close this connection and move on
	if p.idle == closed || p.available.Size() >= p.idle {
		conn.Close()
		return
	}

	// store the connection in our idle connections stack
	p.available.Push(conn)
}
