// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"testing"

	"github.com/shoenig/test/must"
	"github.com/shoenig/test/skip"
)

// nolint:paralleltest
func TestE2E_Set(t *testing.T) {
	skip.CommandUnavailable(t, "memcached")

	c := New(SetServer("localhost:11211"))
	defer func() { _ = c.Close() }()

	err := Set(c, "key1", "value1")
	must.NoError(t, err)

	err = Set(c, "key2", &person{Name: "bob", Age: 34})
	must.NoError(t, err)

	v, verr := Get[string](c, "key1")
	must.NoError(t, verr)
	must.Eq(t, "value1", v)

	v2, verr2 := Get[*person](c, "key2")
	must.NoError(t, verr2)
	must.Eq(t, "bob", v2.Name)
	must.Eq(t, 34, v2.Age)
}
