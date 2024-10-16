// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"math"
	"strconv"
	"testing"

	"github.com/shoenig/test/must"
)

type person struct {
	Name string
	Age  int
}

func TestClient_pick(t *testing.T) {
	t.Parallel()

	t.Run("single", func(t *testing.T) {
		c := New(SetServer("localhost"))

		result := c.pick("foo")
		must.Eq(t, "localhost", result)

		result = c.pick("bar")
		must.Eq(t, "localhost", result)
	})

	t.Run("multi", func(t *testing.T) {
		c := New(
			SetServer("one.local"),
			SetServer("two.local"),
			SetServer("three.local"),
		)

		counts := make(map[string]int)

		for i := 0; i < 1000; i++ {
			key := strconv.Itoa(i)
			result := c.pick(key)
			counts[result]++
		}

		// ensure reasonable distribution
		must.Greater(t, 200, counts["one.local"])
		must.Greater(t, 200, counts["two.local"])
		must.Greater(t, 200, counts["three.local"])
	})
}

func Test_encode(t *testing.T) {
	t.Parallel()

	t.Run("[]byte", func(t *testing.T) {
		b, err := encode([]byte{2, 4, 6, 8})
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("string", func(t *testing.T) {
		b, err := encode("foobar")
		must.NoError(t, err)
		must.SliceLen(t, 6, b)
	})

	t.Run("int8", func(t *testing.T) {
		var i int8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	t.Run("uint8", func(t *testing.T) {
		var i uint8 = 3
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 1, b)
	})

	t.Run("int16", func(t *testing.T) {
		var i int16 = math.MaxInt16
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 2, b)
	})

	t.Run("uint16", func(t *testing.T) {
		var i uint16 = math.MaxUint16
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 2, b)
	})

	t.Run("int32", func(t *testing.T) {
		var i int32 = math.MaxInt32
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("uint32", func(t *testing.T) {
		var i uint32 = math.MaxUint32
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 4, b)
	})

	t.Run("int64", func(t *testing.T) {
		var i int64 = math.MaxInt64
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("uint64", func(t *testing.T) {
		var i uint64 = math.MaxUint64
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("int", func(t *testing.T) {
		var i = math.MaxInt
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("uint", func(t *testing.T) {
		var i uint = math.MaxUint
		b, err := encode(i)
		must.NoError(t, err)
		must.SliceLen(t, 8, b)
	})

	t.Run("struct", func(t *testing.T) {
		p := &person{
			Name: "bob",
			Age:  32,
		}
		b, err := encode(p)
		must.NoError(t, err)
		must.SliceLen(t, 48, b) // sure
	})
}

func Test_decode(t *testing.T) {
	t.Parallel()

	t.Run("[]byte", func(t *testing.T) {
		result, err := decode[[]byte]([]byte{1, 2})
		must.NoError(t, err)
		must.Eq(t, []byte{1, 2}, result)
	})

	t.Run("string", func(t *testing.T) {
		s := []byte("hello")
		result, err := decode[string](s)
		must.NoError(t, err)
		must.Eq(t, "hello", result)
	})

	t.Run("int8", func(t *testing.T) {
		result, err := decode[int8]([]byte{0xfe}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint8", func(t *testing.T) {
		result, err := decode[uint8]([]byte{0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint8, result)
	})

	t.Run("int16", func(t *testing.T) {
		result, err := decode[int16]([]byte{0xfe, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint16", func(t *testing.T) {
		result, err := decode[uint16]([]byte{0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint16, result)
	})

	t.Run("int32", func(t *testing.T) {
		result, err := decode[int32]([]byte{0xfe, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint32", func(t *testing.T) {
		result, err := decode[uint32]([]byte{0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint32, result)
	})

	t.Run("int64", func(t *testing.T) {
		result, err := decode[int64]([]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint64", func(t *testing.T) {
		result, err := decode[uint64]([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint64, result)
	})

	t.Run("int", func(t *testing.T) {
		result, err := decode[int]([]byte{0xfe, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) // little endian
		must.NoError(t, err)
		must.Eq(t, -2, result) // 2's compliment
	})

	t.Run("uint", func(t *testing.T) {
		result, err := decode[uint]([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
		must.NoError(t, err)
		must.Eq(t, math.MaxUint, result)
	})

	t.Run("struct pointer", func(t *testing.T) {
		input, ierr := encode(&person{
			Name: "bob",
			Age:  32,
		})
		must.NoError(t, ierr)
		must.NotNil(t, input)

		result, err := decode[*person](input)
		must.NoError(t, err)
		must.Eq(t, &person{
			Name: "bob",
			Age:  32,
		}, result)
	})

	t.Run("struct value", func(t *testing.T) {
		input, ierr := encode(person{
			Name: "alice",
			Age:  30,
		})
		must.NoError(t, ierr)
		must.NotNil(t, input)

		result, err := decode[person](input)
		must.NoError(t, err)
		must.Eq(t, person{
			Name: "alice",
			Age:  30,
		}, result)
	})
}
