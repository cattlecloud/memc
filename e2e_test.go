// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"testing"
	"time"

	"cattlecloud.net/go/memc/memctest"
	"github.com/shoenig/ignore"
	"github.com/shoenig/test/must"
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
		err := c.Set("mystring", "myvalue")
		must.NoError(t, err)

		var v string
		v, err = c.Get[string]("mystring")
		must.NoError(t, err)
		must.Eq(t, "myvalue", v)
	})

	t.Run("[]byte", func(t *testing.T) {
		err := c.Set("mybytes", []byte{2, 4, 6, 8})
		must.NoError(t, err)

		var v []byte
		v, err = c.Get[[]byte]("mybytes")
		must.NoError(t, err)
		must.Eq(t, []byte{2, 4, 6, 8}, v)
	})

	t.Run("int", func(t *testing.T) {
		err := c.Set("myint", 998877)
		must.NoError(t, err)

		var v int
		v, err = c.Get[int]("myint")
		must.NoError(t, err)
		must.Eq(t, 998877, v)
	})

	t.Run("struct pointer", func(t *testing.T) {
		p := &person{Name: "Seth", Age: 34}
		err := c.Set("myperson_p", p)
		must.NoError(t, err)

		var v *person
		v, err = c.Get[*person]("myperson_p")
		must.NoError(t, err)
		must.Eq(t, &person{Name: "Seth", Age: 34}, v)
	})

	t.Run("struct value", func(t *testing.T) {
		p := person{Name: "Seth", Age: 34}
		err := c.Set("myperson_v", p)
		must.NoError(t, err)

		var v person
		v, err = c.Get[person]("myperson_v")
		must.NoError(t, err)
		must.Eq(t, person{Name: "Seth", Age: 34}, v)
	})
}

func TestE2E_SetGet_expiration(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("hour", func(t *testing.T) {
		err := c.Set("mykey", "myvalue", TTL(1*time.Hour))
		must.NoError(t, err)
	})

	t.Run("months", func(t *testing.T) {
		ttl := 90 * 24 * time.Hour // 3 months
		err := c.Set("mykey", "myvalue", TTL(ttl))
		must.NoError(t, err)
	})
}

func TestE2E_Get_miss(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	_, err := c.Get[string]("missing")
	must.ErrorIs(t, err, ErrCacheMiss)
}

func TestE2E_Delete(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("not found", func(t *testing.T) {
		err := c.Delete("does-not-exist")
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		err = c.Delete("key1")
		must.NoError(t, err)

		err = c.Delete("key1")
		must.ErrorIs(t, err, ErrNotFound)
	})
}

func TestE2E_Add(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Add("key1", "value1")
		must.NoError(t, err)

		v, verr := c.Get[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, v, "value1")
	})

	t.Run("overwrite", func(t *testing.T) {
		err := c.Set("key2", "value2")
		must.NoError(t, err)

		err = c.Add("key2", "value2.b")
		must.ErrorIs(t, err, ErrNotStored)

		v, verr := c.Get[string]("key2")
		must.NoError(t, verr)
		must.Eq(t, v, "value2")
	})
}

func TestE2E_Replace(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		err = c.Replace("key1", "value1.replaced")
		must.NoError(t, err)

		v, verr := c.Get[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, "value1.replaced", v)
	})

	t.Run("not found", func(t *testing.T) {
		err := c.Replace("key-does-not-exist", "value")
		must.ErrorIs(t, err, ErrNotStored)

		_, verr := c.Get[string]("key-does-not-exist")
		must.ErrorIs(t, verr, ErrCacheMiss)
	})
}

func TestE2E_Append(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		err = c.Append("key1", ".appended")
		must.NoError(t, err)

		v, verr := c.Get[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, "value1.appended", v)
	})

	t.Run("not found", func(t *testing.T) {
		err := c.Append("key-does-not-exist", "value")
		must.ErrorIs(t, err, ErrNotStored)
	})
}

func TestE2E_Prepend(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		err = c.Prepend("key1", "prepended.")
		must.NoError(t, err)

		v, verr := c.Get[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, "prepended.value1", v)
	})

	t.Run("not found", func(t *testing.T) {
		err := c.Prepend("key-does-not-exist", "value")
		must.ErrorIs(t, err, ErrNotStored)
	})
}

func TestE2E_Increment(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("unset", func(t *testing.T) {
		_, err := c.Increment("counter-a", 0)
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("negative", func(t *testing.T) {
		err := c.Set("counter-b", "100")
		must.NoError(t, err)

		_, err = c.Increment("counter-b", -2)
		must.ErrorIs(t, err, ErrNegativeInc)
	})

	t.Run("uncountable", func(t *testing.T) {
		err := c.Set("counter-c", "blah")
		must.NoError(t, err)

		_, err = c.Increment("counter-c", 1)
		must.ErrorIs(t, err, ErrNonNumeric)
	})

	t.Run("works", func(t *testing.T) {
		err := c.Set("counter-d", "1000")
		must.NoError(t, err)

		v, verr := c.Increment("counter-d", 2)
		must.NoError(t, verr)
		must.Eq(t, 1002, v)
	})
}

func TestE2E_Decrement(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("unset", func(t *testing.T) {
		_, err := c.Decrement("counter-a", 0)
		must.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("negative", func(t *testing.T) {
		err := c.Set("counter-b", "100")
		must.NoError(t, err)

		_, err = c.Decrement("counter-b", -2)
		must.ErrorIs(t, err, ErrNegativeInc)
	})

	t.Run("uncountable", func(t *testing.T) {
		err := c.Set("counter-c", "blah")
		must.NoError(t, err)

		_, err = c.Decrement("counter-c", 1)
		must.ErrorIs(t, err, ErrNonNumeric)
	})

	t.Run("works", func(t *testing.T) {
		err := c.Set("counter-d", "1000")
		must.NoError(t, err)

		v, verr := c.Decrement("counter-d", 2)
		must.NoError(t, verr)
		must.Eq(t, 998, v)
	})
}

func TestE2E_SetMulti(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	err := c.SetMulti([]*Pair[string, int]{
		{"one", 1},
		{"two", 2},
		{"three", 3},
	})
	must.NoError(t, err)

	one, err1 := c.Get[int]("one")
	must.NoError(t, err1)
	must.Eq(t, 1, one)

	two, err2 := c.Get[int]("two")
	must.NoError(t, err2)
	must.Eq(t, 2, two)

	three, err3 := c.Get[int]("three")
	must.NoError(t, err3)
	must.Eq(t, 3, three)
}

func TestE2E_AddMulti(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	err := c.AddMulti([]*Pair[string, int]{
		{"one", 1},
		{"two", 2},
		{"three", 3},
	})
	must.NoError(t, err)

	one, err1 := c.Get[int]("one")
	must.NoError(t, err1)
	must.Eq(t, 1, one)

	two, err2 := c.Get[int]("two")
	must.NoError(t, err2)
	must.Eq(t, 2, two)

	three, err3 := c.Get[int]("three")
	must.NoError(t, err3)
	must.Eq(t, 3, three)
}

func TestE2E_GetMulti(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	err := c.SetMulti([]*Pair[string, int]{
		{"one", 1},
		{"two", 2},
		{"three", 3},
	})
	must.NoError(t, err)

	results := c.GetMulti[int]([]string{"one", "two", "three"})
	must.Eq(t, []*Pair[int, error]{
		{A: 1, B: nil},
		{A: 2, B: nil},
		{A: 3, B: nil},
	}, results)
}

func TestE2E_GetMulti_missing(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	err := c.SetMulti([]*Pair[string, int]{
		{"one", 1},
		{"three", 3},
	})
	must.NoError(t, err)

	results := c.GetMulti[int]([]string{"one", "two", "three"})
	must.Eq(t, &Pair[int, error]{A: 1, B: nil}, results[0])
	must.Eq(t, &Pair[int, error]{A: 3, B: nil}, results[2])
	must.ErrorIs(t, ErrCacheMiss, results[1].B)
}

func TestE2E_Stats(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	// insert an item
	err := c.Set("mykey", "myvalue", TTL(1*time.Hour))
	must.NoError(t, err)

	s, serr := c.Stats()
	must.NoError(t, serr)

	// spot check a few fields
	must.StrHasPrefix(t, "1.", s.Runtime.Version)
	must.Positive(t, s.Runtime.Threads)
	must.Positive(t, s.Connections.Max)
	must.Positive(t, s.Connections.Current)
	must.One(t, s.Items.Current)
	must.Eq(t, 71, s.Items.Bytes)
}

func TestE2E_StatsSlabs(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	// insert an item
	err := c.Set("mykey", "myvalue", TTL(1*time.Hour))
	must.NoError(t, err)

	s, serr := c.StatsSlabs()
	must.NoError(t, serr)

	// spot check a few fields
	must.Positive(t, s.ActiveSlabs)
	must.Positive(t, s.TotalMalloced)
	must.Positive(t, len(s.Slabs))
	must.Positive(t, s.Slabs[0].ChunkSize)
}

func TestE2E_StatsItems(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	// insert an item
	err := c.Set("mykey", "myvalue", TTL(1*time.Hour))
	must.NoError(t, err)

	data, derr := c.StatsItems()
	must.NoError(t, derr)

	// spot check a few fields
	must.Positive(t, data[0].Class)
	must.Positive(t, data[0].Number)
	must.Positive(t, data[0].MemRequested)
}

func TestE2E_Flush(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		v, verr := c.Get[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, "value1", v)

		err = c.Flush(0)
		must.NoError(t, err)

		_, err = c.Get[string]("key1")
		must.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("empty cache", func(t *testing.T) {
		err := c.Flush(0)
		must.NoError(t, err)
	})
}

func TestE2E_CAS(t *testing.T) {
	t.Parallel()

	address, done := memctest.LaunchTCP(t, nil)
	t.Cleanup(done)

	c := New([]string{address})
	defer ignore.Close(c)

	t.Run("success", func(t *testing.T) {
		err := c.Set("key1", "value1")
		must.NoError(t, err)

		v, cas, verr := c.Gets[string]("key1")
		must.NoError(t, verr)
		must.Eq(t, "value1", v)
		must.Positive(t, uint64(cas))

		err = c.CompareAndSwap("key1", cas, "value1.updated")
		must.NoError(t, err)

		v, err = c.Get[string]("key1")
		must.NoError(t, err)
		must.Eq(t, "value1.updated", v)
	})

	t.Run("conflict", func(t *testing.T) {
		err := c.Set("key2", "original")
		must.NoError(t, err)

		_, cas1, verr := c.Gets[string]("key2")
		must.NoError(t, verr)

		_, _, verr = c.Gets[string]("key2")
		must.NoError(t, verr)

		err = c.CompareAndSwap("key2", cas1, "first-update")
		must.NoError(t, err)

		err = c.CompareAndSwap("key2", cas1, "stale-update")
		must.ErrorIs(t, err, ErrConflict)

		v, err := c.Get[string]("key2")
		must.NoError(t, err)
		must.Eq(t, "first-update", v)
	})

	t.Run("not found", func(t *testing.T) {
		err := c.Set("key3", "value3")
		must.NoError(t, err)

		_, cas, verr := c.Gets[string]("key3")
		must.NoError(t, verr)

		err = c.Delete("key3")
		must.NoError(t, err)

		err = c.CompareAndSwap("key3", cas, "newvalue")
		must.ErrorIs(t, err, ErrNotFound)
	})
}
