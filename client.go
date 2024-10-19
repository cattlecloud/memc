// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"errors"
	"net"
	"regexp"
	"sort"
	"sync"
	"time"
)

type Client struct {
	timeout    time.Duration
	expiration time.Duration

	lock    sync.RWMutex
	servers []string
	conns   []net.Conn
}

func (c *Client) getConn(key string) (*bufio.ReadWriter, error) {
	idx := c.pick(key)

	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	conn := c.conns[idx]
	if conn == nil {
		conn, err = c.open(c.servers[idx])
		c.conns[idx] = conn
	}

	// wrap the connection in a buffer - note that we must now always
	// remember to flush when done writing
	rw := bufio.NewReadWriter(
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
	)

	return rw, err
}

func (c *Client) open(address string) (net.Conn, error) {
	// TODO: unix socket
	// and probably more

	return net.DialTimeout("tcp", address, c.timeout)
}

type ClientOption func(c *Client)

func SetServer(address string) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()

		c.conns = append(c.conns, nil)
		c.servers = append(c.servers, address)
		sort.Strings(c.servers)

		// TODO massive bug; but we should replace all this with a reusable
		// connection pool per address anyway
	}
}

// SetDialTimeout adjusts the amount of time to wait on establishing a TCP
// connection to the memached instance(s).
//
// If unset the default timeout is 5 seconds.
func SetDialTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.timeout = timeout
	}
}

// SetDefaultTTL adjusts the default expiration time of values set into the memcached
// instance(s).
//
// If unset the default expiration TTL is 1 hour.
//
// The expiration time must be more than 1 second, or set to 0 to indicate no
// expiration time (and values stay in the cache indefinitely).
func SetDefaultTTL(expiration time.Duration) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.expiration = expiration
	}
}

const (
	defaultDialTimeout = 5 * time.Second
	defaultExpiration  = 1 * time.Hour
)

func New(opts ...ClientOption) *Client {
	c := new(Client)
	c.timeout = defaultDialTimeout
	c.expiration = defaultExpiration

	for _, opt := range opts {
		opt(c)
	}

	return c
}

var (
	keyRe = regexp.MustCompile(`^[^\s]{1,250}$`)
)

func check(key string) error {
	if !keyRe.MatchString(key) {
		return ErrKeyNotValid
	}
	return nil
}

func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var errs error
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (c *Client) pick(key string) int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if len(c.servers) == 1 {
		return 0
	}

	// compute the server to choose for key
	// deterministic given set of servers and key
	x := byte(37)
	for _, c := range key {
		x ^= byte(c)
	}
	idx := int(int(x) % len(c.servers))

	return idx
}

func seconds(expiration time.Duration) (int, error) {
	if expiration == 0 {
		return 0, nil
	}

	if expiration < 1*time.Second {
		return 0, ErrExpiration
	}

	s := int(expiration.Seconds())
	return s, nil
}
