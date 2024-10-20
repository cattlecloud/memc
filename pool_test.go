// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/shoenig/test/must"
	"noxide.lol/go/stacks"
)

func (p *pool) Equal(o *pool) bool {
	switch {
	case p.address != o.address:
		return false
	case p.timeout != o.timeout:
		return false
	case p.idle != o.idle:
		return false
	case p.available.Size() != o.available.Size():
		return false
	default:
		return true
	}
}

func TestPools_create(t *testing.T) {
	t.Parallel()

	p := new(pools)
	p.create([]string{
		"localhost:2003",
		"localhost:2001",
		"localhost:2002",
	}, 3*time.Second, 4)

	must.SliceEqual(t, []*pool{
		{
			timeout:   3 * time.Second,
			address:   "localhost:2001",
			idle:      4,
			available: stacks.Simple[net.Conn](),
		},
		{
			timeout:   3 * time.Second,
			address:   "localhost:2002",
			idle:      4,
			available: stacks.Simple[net.Conn](),
		},
		{
			timeout:   3 * time.Second,
			address:   "localhost:2003",
			idle:      4,
			available: stacks.Simple[net.Conn](),
		},
	}, p.servers)
}

func TestPools_pick(t *testing.T) {
	t.Parallel()

	t.Run("single", func(t *testing.T) {
		p := &pools{
			servers: []*pool{
				{address: "one.local"},
			},
		}

		result := p.pick("foo")
		must.Eq(t, 0, result)

		result = p.pick("bar")
		must.Eq(t, 0, result)
	})

	t.Run("multi", func(t *testing.T) {
		p := &pools{
			servers: []*pool{
				{address: "one.local"},
				{address: "two.local"},
				{address: "three.local"},
			},
		}

		counts := make(map[int]int)

		for i := 0; i < 1000; i++ {
			key := strconv.Itoa(i)
			result := p.pick(key)
			counts[result]++
		}

		// ensure reasonable distribution
		must.Greater(t, 200, counts[0])
		must.Greater(t, 200, counts[1])
		must.Greater(t, 200, counts[2])
	})
}

func TestPools_close(t *testing.T) {
	t.Parallel()

	p := new(pools)
	p.create([]string{
		"localhost:2001",
		"localhost:2002",
	}, 3*time.Second, 4)

	p.close()

	_, _, err := p.get("anything")
	must.ErrorIs(t, err, ErrClientClosed)
}

func listenTCP(t *testing.T) (string, net.Listener) {
	port := ports.One()
	address := fmt.Sprintf("localhost:%d", port)
	l, err := net.Listen("tcp", address)
	must.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	return address, l
}

func TestPool_get(t *testing.T) {
	t.Parallel()

	address, _ := listenTCP(t)

	p := &pool{
		timeout:   1 * time.Second,
		address:   address,
		idle:      2,
		available: stacks.Simple[net.Conn](),
	}

	con, _, err := p.get()
	must.NoError(t, err)
	must.Close(t, con)
}

func TestPool_discard(t *testing.T) {
	t.Parallel()

	address, _ := listenTCP(t)

	p := &pool{
		timeout:   1 * time.Second,
		address:   address,
		idle:      1,
		available: stacks.Simple[net.Conn](),
	}

	con1, _, err1 := p.get()
	must.NoError(t, err1)

	// puts the connection on the idle stack
	p.discard(con1)
	must.Eq(t, 1, p.available.Size())

	con2, _, err2 := p.get()
	must.NoError(t, err2)
	must.Eq(t, 0, p.available.Size())

	con3, _, err3 := p.get()
	must.NoError(t, err3)

	p.discard(con2) // remains open, idle
	p.discard(con3) // closed, not needed

	must.Error(t, con3.Close())   // should be closed already
	must.NoError(t, con2.Close()) // should still be open / idle
}
