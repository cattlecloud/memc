// Copyright (c) The Noxide Project Authors
// SPDX-License-Identifier: BSD-3-Clause

package memc

import "errors"

var (
	ErrCacheMiss   = errors.New("cache miss")
	ErrKeyNotValid = errors.New("key is not valid")
)

func Set[T any](c *Client, key string, item T) error {
	return nil
}

func Get[T any](c *Client, key string) (T, error) {
	var empty T
	return empty, nil
}
