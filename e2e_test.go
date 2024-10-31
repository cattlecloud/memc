// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"testing"
	"time"

	"github.com/shoenig/ignore"
	"github.com/shoenig/test/must"
	"noxide.lol/go/memc/memctest"
)

// Examples using netcat
//
// echo -n -e "set key 0 300 3\r\nval\r\n" | nc localhost 11211
//
// echo -n -e "delete key\r\n" | nc localhost 11211

func TestE2E_SetGet_simple(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("string", func(t *testing.T) {
		err := Set(c, "mystring", "myvalue")
		must.NoError(t, err)

		var v string
		v, err = Get[string](c, "mystring")
		must.NoError(t, err)
		must.Eq(t, "myvalue", v)
	})

	t.Run("[]byte", func(t *testing.T) {
		err := Set(c, "mybytes", []byte{2, 4, 6, 8})
		must.NoError(t, err)

		var v []byte
		v, err = Get[[]byte](c, "mybytes")
		must.NoError(t, err)
		must.Eq(t, []byte{2, 4, 6, 8}, v)
	})

	t.Run("int", func(t *testing.T) {
		err := Set(c, "myint", 998877)
		must.NoError(t, err)

		var v int
		v, err = Get[int](c, "myint")
		must.NoError(t, err)
		must.Eq(t, 998877, v)
	})

	t.Run("struct pointer", func(t *testing.T) {
		p := &person{Name: "Seth", Age: 34}
		err := Set(c, "myperson_p", p)
		must.NoError(t, err)

		var v *person
		v, err = Get[*person](c, "myperson_p")
		must.NoError(t, err)
		must.Eq(t, &person{Name: "Seth", Age: 34}, v)
	})

	t.Run("struct value", func(t *testing.T) {
		p := person{Name: "Seth", Age: 34}
		err := Set(c, "myperson_v", p)
		must.NoError(t, err)

		var v person
		v, err = Get[person](c, "myperson_v")
		must.NoError(t, err)
		must.Eq(t, person{Name: "Seth", Age: 34}, v)
	})
}

func Test_SetGet_expiration(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("hour", func(t *testing.T) {
		err := Set(c, "mykey", "myvalue", TTL(1*time.Hour))
		must.NoError(t, err)
	})
}

func Test_Get_miss(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	_, err := Get[string](c, "missing")
	must.ErrorIs(t, err, ErrCacheMiss)
}

func Test_Delete(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("not found", func(t *testing.T) {
		err := Delete(c, "does-not-exist")
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("success", func(t *testing.T) {
		err := Set(c, "key1", "value1")
		must.NoError(t, err)

		err = Delete(c, "key1")
		must.NoError(t, err)

		err = Delete(c, "key1")
		must.ErrorIs(t, err, ErrNotFound)
	})
}

func Test_Add(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := Add(c, "key1", "value1")
		must.NoError(t, err)

		v, verr := Get[string](c, "key1")
		must.NoError(t, verr)
		must.Eq(t, v, "value1")
	})

	t.Run("overwrite", func(t *testing.T) {
		err := Set(c, "key2", "value2")
		must.NoError(t, err)

		err = Add(c, "key2", "value2.b")
		must.ErrorIs(t, err, ErrNotStored)

		v, verr := Get[string](c, "key2")
		must.NoError(t, verr)
		must.Eq(t, v, "value2")
	})
}

func Test_Increment(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("unset", func(t *testing.T) {
		_, err := Increment(c, "counter-a", 0)
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("negative", func(t *testing.T) {
		err := Set(c, "counter-b", "100")
		must.NoError(t, err)

		_, err = Increment(c, "counter-b", -2)
		must.ErrorIs(t, err, ErrNegativeInc)
	})

	t.Run("uncountable", func(t *testing.T) {
		err := Set(c, "counter-c", "blah")
		must.NoError(t, err)

		_, err = Increment(c, "counter-c", 1)
		must.ErrorIs(t, err, ErrNonNumeric)
	})

	t.Run("works", func(t *testing.T) {
		err := Set(c, "counter-d", "1000")
		must.NoError(t, err)

		v, verr := Increment(c, "counter-d", 2)
		must.NoError(t, verr)
		must.Eq(t, 1002, v)
	})
}

func Test_Decrement(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("unset", func(t *testing.T) {
		_, err := Decrement(c, "counter-a", 0)
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("negative", func(t *testing.T) {
		err := Set(c, "counter-b", "100")
		must.NoError(t, err)

		_, err = Decrement(c, "counter-b", -2)
		must.ErrorIs(t, err, ErrNegativeInc)
	})

	t.Run("uncountable", func(t *testing.T) {
		err := Set(c, "counter-c", "blah")
		must.NoError(t, err)

		_, err = Decrement(c, "counter-c", 1)
		must.ErrorIs(t, err, ErrNonNumeric)
	})

	t.Run("works", func(t *testing.T) {
		err := Set(c, "counter-d", "1000")
		must.NoError(t, err)

		v, verr := Decrement(c, "counter-d", 2)
		must.NoError(t, verr)
		must.Eq(t, 998, v)
	})
}
