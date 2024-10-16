// Package memc provides a modern, efficient memcached client for Go.
package memc

import (
	"bytes"
	"encoding/binary"
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
	case int16:
		b := make([]byte, 2)
		i := item.(int16)
		binary.LittleEndian.PutUint16(b, uint16(i))
		return b, nil
	case uint16:
		b := make([]byte, 2)
		i := item.(uint16)
		binary.LittleEndian.PutUint16(b, i)
		return b, nil
	case int32:
		b := make([]byte, 4)
		i := item.(int32)
		binary.LittleEndian.PutUint32(b, uint32(i))
		return b, nil
	case uint32:
		b := make([]byte, 4)
		i := item.(uint32)
		binary.LittleEndian.PutUint32(b, i)
		return b, nil
	case int64:
		b := make([]byte, 8)
		i := item.(int64)
		binary.LittleEndian.PutUint64(b, uint64(i))
		return b, nil
	case uint64:
		b := make([]byte, 8)
		i := item.(uint64)
		binary.LittleEndian.PutUint64(b, i)
		return b, nil
	case int:
		b := make([]byte, 8)
		i := item.(int)
		binary.LittleEndian.PutUint64(b, uint64(i))
		return b, nil
	case uint:
		b := make([]byte, 8)
		i := item.(uint)
		binary.LittleEndian.PutUint64(b, uint64(i))
		return b, nil
	default:
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(item)
		return buf.Bytes(), err
	}
}

func decode[T any](b []byte) (T, error) {
	var result T
	switch any(result).(type) {
	case []byte:
		tmp := any(b).(T)
		return tmp, nil
	case string:
		s := string(b)
		tmp := any(s).(T)
		return tmp, nil
	case int8:
		i := int8(b[0])
		tmp := any(i).(T)
		return tmp, nil
	case uint8:
		i := b[0]
		tmp := any(i).(T)
		return tmp, nil
	case int16:
		i := int16(binary.LittleEndian.Uint16(b))
		tmp := any(i).(T)
		return tmp, nil
	case uint16:
		i := binary.LittleEndian.Uint16(b)
		tmp := any(i).(T)
		return tmp, nil
	case int32:
		i := int32(binary.LittleEndian.Uint32(b))
		tmp := any(i).(T)
		return tmp, nil
	case uint32:
		i := binary.LittleEndian.Uint32(b)
		tmp := any(i).(T)
		return tmp, nil
	case int64:
		i := int64(binary.LittleEndian.Uint64(b))
		tmp := any(i).(T)
		return tmp, nil
	case uint64:
		i := binary.LittleEndian.Uint64(b)
		tmp := any(i).(T)
		return tmp, nil
	case int:
		i := int(binary.LittleEndian.Uint64(b))
		tmp := any(i).(T)
		return tmp, nil
	case uint:
		i := uint(binary.LittleEndian.Uint64(b))
		tmp := any(i).(T)
		return tmp, nil
	default:
		buf := bytes.NewBuffer(b)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&result)
		return result, err
	}
}
