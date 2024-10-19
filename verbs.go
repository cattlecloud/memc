// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

var (
	ErrCacheMiss   = errors.New("memc: cache miss")
	ErrKeyNotValid = errors.New("memc: key is not valid")
	ErrNotStored   = errors.New("memc: item not stored")
	ErrNotFound    = errors.New("memc: item not found")
	ErrConflict    = errors.New("memc: CAS conflict")
)

func Set[T any](c *Client, key string, item T) error {
	if err := check(key); err != nil {
		return err
	}

	rw, cerr := c.getConn(key)
	if cerr != nil {
		return cerr
	}

	flags := 0
	expiration := 0
	bs, nerr := encode(item)
	if nerr != nil {
		return nerr
	}

	// write the header components
	if _, err := fmt.Fprintf(
		rw,
		"set %s %d %d %d\r\n",
		key, flags, expiration, len(bs),
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
	_ = key
	_ = item
	return nil
}

func Touch(c *Client, key string) error {
	if err := check(key); err != nil {
		return err
	}

	_ = c
	return nil
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
	return nil
}
