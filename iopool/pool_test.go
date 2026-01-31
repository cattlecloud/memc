// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package iopool

import (
	"errors"
	"fmt"
	"testing"

	"github.com/shoenig/test/must"
)

func TestBuffer_SetHealth(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		b := newBuffer(nil)
		must.False(t, b.failure.Load())
	})

	t.Run("nil", func(t *testing.T) {
		b := newBuffer(nil)
		b.SetHealth(nil)
		must.False(t, b.failure.Load())
	})

	t.Run("error", func(t *testing.T) {
		b := newBuffer(nil)
		b.SetHealth(errors.New("oops"))
		must.True(t, b.failure.Load())
	})
}

func TestPool_get(t *testing.T) {
	t.Parallel()

	t.Run("closed", func(t *testing.T) {
		p := newPool("10.0.0.1", 1)
		p.openf = mockConnections(
			newMockConn(nil, nil),
		)
		p.idle = closed
		c, err := p.get()
		must.ErrorIs(t, err, ErrClientClosed)
		must.Nil(t, c)
	})

	t.Run("normal", func(t *testing.T) {
		p := newPool("10.0.0.1", 1)
		p.openf = mockConnections(
			newMockConn(nil, nil),
		)
		c, err := p.get()
		must.NoError(t, err)
		must.NotNil(t, c)
	})

	t.Run("second", func(t *testing.T) {
		p := newPool("10.0.0.1", 1)
		p.openf = mockConnections(
			newMockConn(nil, nil),
			newMockConn(nil, nil),
		)

		c, err := p.get()
		must.NoError(t, err)
		must.NotNil(t, c)

		c, err = p.get()
		must.NoError(t, err)
		must.NotNil(t, c)
	})
}

func TestPool_free(t *testing.T) {
	t.Parallel()

	t.Run("closed", func(t *testing.T) {
		p := newPool("10.0.0.1", 1)
		p.openf = mockConnections(
			newMockConn(nil, nil),
		)

		c, err := p.get()
		must.NoError(t, err)

		p.close()
		p.free(c)
		must.Empty(t, p.available)
	})

	t.Run("full", func(t *testing.T) {
		p := newPool("10.0.0.1", 2)
		p.openf = mockConnections(
			newMockConn(nil, nil),
			newMockConn(nil, nil),
			newMockConn(nil, nil),
		)

		c1, err1 := p.get()
		must.NoError(t, err1)

		c2, err2 := p.get()
		must.NoError(t, err2)

		c3, err3 := p.get()
		must.NoError(t, err3)

		// totally empty
		must.Eq(t, 0, p.available.Size())

		// one idle connection
		p.free(c1)
		must.Eq(t, 1, p.available.Size())

		// two idle connections
		p.free(c2)
		must.Eq(t, 2, p.available.Size())

		// throw away overflow connection
		p.free(c3)
		must.Eq(t, 2, p.available.Size())
	})

	t.Run("failure", func(t *testing.T) {
		p := newPool("10.0.0.1", 2)
		p.openf = mockConnections(
			newMockConn(nil, nil),
		)

		c, err := p.get()
		must.NoError(t, err)

		c.SetHealth(errors.New("oops"))

		// discard connection that had a failure
		must.Empty(t, p.available)
		p.free(c)
		must.Empty(t, p.available)
	})
}

func TestCollection_pick_distribution(t *testing.T) {
	t.Parallel()

	c := &Collection{
		pools: []*pool{
			{}, {}, {},
		},
	}

	counts := make(map[int]int)

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		idx := c.pick(key)
		counts[idx]++
	}

	must.Greater(t, 200, counts[0])
	must.Greater(t, 200, counts[1])
	must.Greater(t, 200, counts[2])
}

func TestCollection_GetReturn(t *testing.T) {
	t.Parallel()

	p := newPool("10.0.0.1", 1)
	p.openf = mockConnections(
		newMockConn(nil, nil),
	)

	c := &Collection{
		pools: []*pool{p},
	}

	conn, err := c.Get("abc123")
	must.NoError(t, err)

	c.Return("abc123", conn)
}

func TestCollection_GetCloseReturn(t *testing.T) {
	t.Parallel()

	p := newPool("10.0.0.1", 1)
	p.openf = mockConnections(
		newMockConn(nil, nil),
	)

	c := &Collection{
		pools: []*pool{p},
	}

	conn, err := c.Get("abc123")
	must.NoError(t, err)

	err = c.Close()
	must.NoError(t, err)

	c.Return("abc123", conn)
}
