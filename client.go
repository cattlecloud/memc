// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"regexp"
	"sync"
	"time"

	"cattlecloud.net/go/memc/iopool"
)

// A Client is used for making network requests to memcached instances.
//
// Use the package functions Set, Get, Delete, etc. by providing this Client to
// manage data in memcached.
type Client struct {
	timeout    time.Duration
	expiration time.Duration
	idle       int
	now        func() time.Time

	lock  sync.Mutex
	addrs []string
	pools *iopool.Collection
}

func (c *Client) getConn(key string) (*iopool.Buffer, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.pools.Get(key)
}

func (c *Client) setConn(key string, conn *iopool.Buffer) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.pools.Return(key, conn)
}

type ClientOption func(c *Client)

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

// ClockFunc is a function that returns the current time.
//
// Normally this should just be the time.Now function.
type ClockFunc func() time.Time

// SetClock sets the ClockFunc used for getting the current time.
//
// If unset the default is to use the time.Now function.
//
// Note this should typically only be set in testing.
func SetClock(f ClockFunc) ClientOption {
	return func(c *Client) {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.now = f
	}
}

const (
	defaultDialTimeout = 5 * time.Second
	defaultExpiration  = 1 * time.Hour
	defaultIdleCount   = 1
)

// New creates a new Client capable of sharding across the given set of
// instances and pooling connections to each instance.
//
// Certain behaviors can be configured by specifying one or more ClientOption
// options.
func New(instances []string, opts ...ClientOption) *Client {
	c := new(Client)
	c.addrs = instances
	c.timeout = defaultDialTimeout
	c.expiration = defaultExpiration
	c.idle = defaultIdleCount
	c.now = time.Now

	for _, opt := range opts {
		opt(c)
	}

	c.pools = iopool.New(c.addrs, c.idle)
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

	return c.pools.Close()
}

// seconds returns the number of seconds until expiration, unless the
// expiration is more than 30 days (2_592_000 seconds), in which case the
// absolute timestamp is used and expected by the memcached instance
func (c *Client) seconds(expiration time.Duration) (int, error) {
	switch {
	case expiration == 0:
		return 0, nil
	case expiration < 1*time.Second:
		return 0, ErrExpiration
	case expiration > 2_592_000*time.Second:
		unix := c.now()
		later := unix.Add(expiration)
		s := int(later.Unix())
		return s, nil
	default:
		s := int(expiration.Seconds())
		return s, nil
	}
}

func (c *Client) do(key string, f func(*iopool.Buffer) error) error {
	conn, err := c.getConn(key)
	if err != nil {
		return err
	}
	err = f(conn)
	conn.SetHealth(err)
	c.setConn(key, conn)
	return err
}
