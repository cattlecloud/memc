// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

// Countable represents types that work with Increment and Decrement operations.
//
// Note: memcached does not allow negative values for either operation.
type Countable interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~int
}

func encode(item any) ([]byte, error) {
	switch v := item.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int8:
		b := byte(v)
		return []byte{b}, nil
	case uint8:
		b := byte(v)
		return []byte{b}, nil
	case int16:
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(v))
		return b, nil
	case uint16:
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, v)
		return b, nil
	case int32:
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(v))
		return b, nil
	case uint32:
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, v)
		return b, nil
	case int64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		return b, nil
	case uint64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, v)
		return b, nil
	case int:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
		return b, nil
	case uint:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(v))
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
