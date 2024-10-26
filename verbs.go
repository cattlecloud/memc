// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
	ErrNegativeInc  = errors.New("memc: increment delta must be non-negative")
	ErrNonNumeric   = errors.New("memc: cannot increment non-numeric value")
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

// Set will store the item using the given key, possibly overwriting any
// existing data. New items are at the top of the LRU.
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
		default:
			return fmt.Errorf("memc: unexpected response to set: %q", string(line))
		}
	})
}

// Add will store the item using the given key, but only if no item currently
// exists. New items are at the top of the LRU.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
//
// One or more Option(s) may be applied to configure things such as the
// value expiration TTL or its associated flags.
func Add[T any](c *Client, key string, item T, opts ...Option) error {
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
			"add %s %d %d %d\r\n",
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
		default:
			return fmt.Errorf("memc: unexpected response to set: %q", string(line))
		}
	})
}

// Get the value associated with the given key.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
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
		return nil, unexpected(b)
	}

	return payload, err
}

// Delete will remove the value associated with key from memcached.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
func Delete(c *Client, key string) error {
	if err := check(key); err != nil {
		return err
	}

	return c.do(key, func(conn *iopool.Buffer) error {
		// write the header components
		if _, err := fmt.Fprintf(
			conn,
			"delete %s\r\n",
			key,
		); err != nil {
			return err
		}

		// flush the buffer
		if err := conn.Flush(); err != nil {
			return err
		}

		line, lerr := conn.ReadSlice('\n')
		if lerr != nil {
			return lerr
		}

		switch string(line) {
		case "DELETED\r\n":
			return nil
		case "NOT_FOUND\r\n":
			return ErrNotFound
		default:
			return unexpected(line)
		}
	})
}

// Increment will increment the value associated with the given key by delta.
//
// Note: the value must be an ASCII integer. It must have been initially stored
// as a string value, e.g. by using Set. The delta value must be positive.
//
//	Set(client, "counter", "100")
//	Increment(client, "counter", 1) // counter = 101
func Increment[T Countable](c *Client, key string, delta T) (T, error) {
	if err := check(key); err != nil {
		return T(0), err
	}

	if delta < 0 {
		return T(0), ErrNegativeInc
	}

	var result T

	err := c.do(key, func(conn *iopool.Buffer) error {
		// write the header components
		if _, err := fmt.Fprintf(
			conn,
			"incr %s %d\r\n",
			key, delta,
		); err != nil {
			return err
		}

		// flush the buffer
		if err := conn.Flush(); err != nil {
			return err
		}

		// read the response
		line, lerr := conn.ReadSlice('\n')
		if lerr != nil {
			return lerr
		}

		// check for error response
		s := string(line)
		switch {
		case s == "NOT_FOUND\r\n":
			return ErrNotFound
		case strings.Contains(s, "cannot increment or decrement non-numeric value"):
			return ErrNonNumeric
		}

		// parse response as the resulting value
		s = strings.TrimSpace(s)
		u, uerr := strconv.ParseUint(s, 10, 64)
		if uerr != nil {
			return unexpected(line)
		}

		// recast to value type
		result = T(u)

		return nil
	})

	return result, err
}

// Decrement will decrement the value associated with the given key by delta.
//
// Note: the value must be an ASCII integer. It must have been initially stored
// as a string value, e.g. by using Set. The delta value must be positive.
//
//	Set(client, "counter", "100")
//	Decrement(client, "counter", 1) // counter = 99
func Decrement[T Countable](c *Client, key string, delta T) (T, error) {
	if err := check(key); err != nil {
		return T(0), err
	}

	if delta < 0 {
		return T(0), ErrNegativeInc
	}

	var result T

	err := c.do(key, func(conn *iopool.Buffer) error {
		// write the header components
		if _, err := fmt.Fprintf(
			conn,
			"decr %s %d\r\n",
			key, delta,
		); err != nil {
			return err
		}

		// flush the buffer
		if err := conn.Flush(); err != nil {
			return err
		}

		// read the response
		line, lerr := conn.ReadSlice('\n')
		if lerr != nil {
			return lerr
		}

		// check for error response
		s := string(line)
		switch {
		case s == "NOT_FOUND\r\n":
			return ErrNotFound
		case strings.Contains(s, "cannot increment or decrement non-numeric value"):
			return ErrNonNumeric
		}

		// parse response as the resulting value
		s = strings.TrimSpace(s)
		u, uerr := strconv.ParseUint(s, 10, 64)
		if uerr != nil {
			return unexpected(line)
		}

		// recast to value type
		result = T(u)

		return nil
	})

	return result, err
}

func unexpected(response []byte) error {
	return fmt.Errorf(
		"unexpected response from memcached %q",
		string(response),
	)
}
