// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"
)

var (
	ErrCacheMiss    = errors.New("memc: cache miss")
	ErrKeyNotValid  = errors.New("memc: key is not valid")
	ErrNotStored    = errors.New("memc: item not stored")
	ErrNotFound     = errors.New("memc: item not found")
	ErrConflict     = errors.New("memc: CAS conflict")
	ErrExpiration   = errors.New("memc: expiration ttl is not valid")
	ErrClientClosed = errors.New("memc: client has been closed")
)

// Options contains configuration parameters that may be applied when executing
// a verb like Get, Set, etc.
type Options struct {
	expiration time.Duration
	flags      int
}

// Option to apply when executing a verb like Get, Set, etc.
type Option func(o *Options)

// TTL applies the given expiration time to set on the value being set.
//
// The expiration must be greater than 1 second, or 0, indicating the value will
// not expire automatically.
func TTL(expiration time.Duration) Option {
	return func(o *Options) {
		o.expiration = expiration
	}
}

// Flags applies the given flags on the value being set.
func Flags(flags int) Option {
	return func(o *Options) {
		o.flags = flags
	}
}

// Set the key to contain the value of item.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
//
// One or more Option(s) may be applied to configure things such as the
// value expiration TTL or its associated flags.
func Set[T any](c *Client, key string, item T, opts ...Option) error {
	if err := check(key); err != nil {
		return err
	}

	options := &Options{
		expiration: c.expiration,
		flags:      0,
	}

	for _, opt := range opts {
		opt(options)
	}

	rw, cerr := c.getConn(key)
	if cerr != nil {
		return cerr
	}

	bs, nerr := encode(item)
	if nerr != nil {
		return nerr
	}

	expiration, serr := seconds(options.expiration)
	if serr != nil {
		return serr
	}

	// write the header components
	if _, err := fmt.Fprintf(
		rw,
		"set %s %d %d %d\r\n",
		key, options.flags, expiration, len(bs),
	); err != nil {
		return err
	}

	// write the payload
	if _, err := rw.Write(bs); err != nil {
		return err
	}

	// write clrf
	if _, err := io.WriteString(rw, "\r\n"); err != nil {
		return err
	}

	// flush the buffer
	if err := rw.Flush(); err != nil {
		return err
	}

	// read response
	line, lerr := rw.ReadSlice('\n')
	if lerr != nil {
		return lerr
	}

	switch string(line) {
	case "STORED\r\n":
		return nil
	case "NOT_STORED\r\n":
		return ErrNotStored
	case "EXISTS\r\n":
		return ErrConflict
	case "NOT_FOUND\r\n":
		return ErrCacheMiss
	default:
		return fmt.Errorf("memc: unexpected response to set: %q", string(line))
	}
}

func Add[T any](c *Client, key string, item T) error {
	if err := check(key); err != nil {
		return err
	}

	_ = c
	_ = item

	panic("not yet implemented")
}

func Touch(c *Client, key string) error {
	if err := check(key); err != nil {
		return err
	}
	_ = c

	panic("not yet implemented")
}

func Get[T any](c *Client, key string) (T, error) {
	var empty T

	if err := check(key); err != nil {
		return empty, err
	}

	rw, cerr := c.getConn(key)
	if cerr != nil {
		return empty, cerr
	}

	// write the header components
	if _, err := fmt.Fprintf(rw, "get %s\r\n", key); err != nil {
		return empty, err
	}

	if err := rw.Flush(); err != nil {
		return empty, err
	}

	payload, perr := getPayload(rw.Reader)
	if perr != nil {
		return empty, perr
	}

	return decode[T](payload)
}

func getPayload(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	// TODO: does not handle CAS value for now
	expect := "VALUE %s %d %d\r\n"
	var (
		key   string
		flags int
		size  int
	)

	// scan the header line, giving us a payload size
	if _, err = fmt.Sscanf(string(b), expect, &key, &flags, &size); err != nil {
		return nil, err
	}

	// read the data into our payload
	payload := make([]byte, size+2) // including \r\n
	if _, err = io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	payload = payload[0:size] // chop \r\n

	// read the trailing line ("END\r\n")
	b, err = r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if string(b) != "END\r\n" {
		return nil, fmt.Errorf("unexpected response from memcache %q", string(b))
	}

	return payload, err
}

func Delete(c *Client, key string) error {
	if err := check(key); err != nil {
		return err
	}

	_ = c

	panic("not yet implemented")
}
