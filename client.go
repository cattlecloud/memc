// Package memc provides a modern, efficient memcached client for Go.
package memc

import (
	"bytes"
	"encoding/gob"
)

type Client struct {
	servers []string
}

type ClientOption func(c *Client)

func SetServer(address string) ClientOption {
	return func(c *Client) {
		c.servers = append(c.servers, address)
	}
}

func New(opts ...ClientOption) *Client {
	c := new(Client)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func encode(item any) ([]byte, error) {
	switch item.(type) {
	case []byte:
		return item.([]byte), nil
	case string:
		return []byte(item.(string)), nil
	case int8:
		b := byte(item.(int8))
		return []byte{b}, nil
	case uint8:
		b := byte(item.(uint8))
		return []byte{b}, nil
	case int16, uint16:
	case int32, uint32:
	case int64, uint64:
	case int, uint:
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(item)
	return buf.Bytes(), err
}

func decode[T any](b []byte, dest any) (T, error) {
	switch dest.(type) {
	case []byte:
		return dest.(T), nil
	}
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(dest)
	return dest.(T), err
}
