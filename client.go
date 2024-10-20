// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"regexp"
	"sync"
	"time"
)

// A Client is used for making network requests to memcached instances.
//
// Use the package functions Set, Get, Delete, etc. by providing this Client to
// manage data in memcached.
type Client struct {
	timeout    time.Duration
	expiration time.Duration
	idle       int

	lock  sync.Mutex
	addrs []string
	pools *pools
}

func (c *Client) getConn(key string) (*bufio.ReadWriter, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	conn, done, err := c.pools.get(key)
	if err != nil {
		return nil, err
	}
	defer done()

	// wrap the connection in a buffer - note that we must now always
	// remember to flush when done writing
	rw := bufio.NewReadWriter(
		bufio.NewReader(conn),
		bufio.NewWriter(conn),
	)

	return rw, err
}

type ClientOption func(c *Client)

// SetServer appends the given server address to the list of memcached instances
// available for storing data.
func SetServer(address string) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()

		c.addrs = append(c.addrs, address)
	}
}

// SetIdleConnections adjusts the maximum number of idle connections to maintain
// for each memcached instance.
//
// If unset the default idle connection limit is 1.
//
// Note that idle connections are created on demand, not at startup.
func SetIdleConnections(count int) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()

		c.idle = count
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
	c.idle = 1

	for _, opt := range opts {
		opt(c)
	}

	c.pools = new(pools)
	c.pools.create(c.addrs, c.timeout, c.idle)
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

// Close will close all idle connections and prevent existing connections from
// becoming idle. Future use of the Client will fail.
func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.pools.close()
	return nil
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
