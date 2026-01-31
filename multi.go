// Copyright CattleCloud LLC 2025, 2026
// SPDX-License-Identifier: BSD-3-Clause

package memc

import "errors"

// A Pair associates two elements.
type Pair[T, U any] struct {
	A T
	B U
}

// SetMulti will store each item in items using the item's associated key,
// possibly overwritting any existing data. New items are at the top of the
// LRU.
//
// Errors are accumulated using errors.Join.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
//
// One or more Option(s) may be applied to configure things such as the
// value expiration TTL or its associated flags.
func SetMulti[T any](c *Client, items []*Pair[string, T], opts ...Option) error {
	var errs []error
	for _, item := range items {
		if err := Set(c, item.A, item.B, opts...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// AddMulti will store each item in items using the item's associated key,
// but only if the item does not currently exist. New items are at the top of
// the LRU.
//
// Errors are accumulated using errors.Join.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
//
// One or more Option(s) may be applied to configure things such as the
// value expiration TTL or its associated flags.
func AddMulti[T any](c *Client, items []*Pair[string, T], opts ...Option) error {
	var errs []error
	for _, item := range items {
		if err := Add(c, item.A, item.B, opts...); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Get the values associated with the given keys. One Pair[T, error] return
// value for each of the given keys, in the same order.
//
// Uses Client c to connect to a memcached instance, and automatically handles
// connection pooling and reuse.
func GetMulti[T any](c *Client, keys []string) []*Pair[T, error] {
	results := make([]*Pair[T, error], 0, len(keys))
	for _, key := range keys {
		v, err := Get[T](c, key)
		if err != nil {
			results = append(results, &Pair[T, error]{B: err})
		} else {
			results = append(results, &Pair[T, error]{A: v})
		}
	}
	return results
}
