// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"testing"

	"cattlecloud.net/go/memc/memctest"
	"github.com/shoenig/ignore"
	"github.com/shoenig/test/must"
)

func TestUnix_simple(t *testing.T) {
	t.Parallel()

	socket, done := memctest.LaunchUDS(t, nil)
	t.Cleanup(done)

	c := New([]string{socket})
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
