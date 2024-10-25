// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package iopool

import (
	"sync"
)

type mockConn struct {
	lock       *sync.Mutex
	sequence   int
	setReads   []string
	expWrites  []string
	errOnRead  error
	errOnWrite error
	errOnClose error
}

func (mc *mockConn) Read([]byte) (int, error) {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	return 0, mc.errOnRead
}

func (mc *mockConn) Write([]byte) (int, error) {
	return 0, mc.errOnWrite
}

func (mc *mockConn) Close() error {
	return mc.errOnClose
}

func newMockConn(reads, writes []string) *mockConn {
	return &mockConn{
		setReads:  reads,
		expWrites: writes,
	}
}

func mockConnections(connections ...*mockConn) func(string) (Connection, error) {
	i := 0
	return func(string) (Connection, error) {
		next := connections[i]
		i++
		next.sequence = i
		return next, nil
	}
}
