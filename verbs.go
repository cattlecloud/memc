// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"noxide.lol/go/memc/iopool"
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

	return c.do(key, func(conn *iopool.Buffer) error {
		encoding, encerr := encode(item)
		if encerr != nil {
			return encerr
		}

		expiration, experr := seconds(options.expiration)
		if experr != nil {
			return experr
		}

		// write the header components
		if _, err := fmt.Fprintf(
			conn,
			"set %s %d %d %d\r\n",
			key, options.flags, expiration, len(encoding),
		); err != nil {
			return err
		}

		// write the payload
		if _, err := conn.Write(encoding); err != nil {
			return err
		}

		// write clrf
		if _, err := io.WriteString(conn, "\r\n"); err != nil {
			return err
		}

		// flush the buffer
		if err := conn.Flush(); err != nil {
			return err
		}

		// read response
		line, lerr := conn.ReadSlice('\n')
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
	})
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

// Get the value associated with the given key.
func Get[T any](c *Client, key string) (T, error) {
	var result T

	if err := check(key); err != nil {
		return result, err
	}

	err := c.do(key, func(conn *iopool.Buffer) error {
		// write the header components
		if _, err := fmt.Fprintf(conn, "get %s\r\n", key); err != nil {
			return err
		}

		// flush the connection, forcing bytes over the wire
		if err := conn.Flush(); err != nil {
			return err
		}

		// read the response payload
		payload, err := getPayload(conn.Reader)
		if err != nil {
			return err
		}

		result, err = decode[T](payload)
		return err
	})

	return result, err
}

func getPayload(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}

	// key was not found, is a cache miss
	if string(b) == "END\r\n" {
		return nil, ErrCacheMiss
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
