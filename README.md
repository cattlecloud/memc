# memc

`memc` is a modern and generics enabled memcached client library for Go.

_requires go1.23+_

October 2024:
(!) This package is very new and may contain bugs and missing features.

### Getting Started

The `memc` package can be added to a Go project with `go get`.

```shell
go get cattlecloud.net/go/memc@latest
```

```go
import "cattlecloud.net/go/memc"
```

### Examples

##### Setting a value in memcached.

```go
client := memc.New(
  []string{"localhost:11211"},
)

err := memc.Set(client, "my/key/name", "some_value")
```

Note that the `memc` library can handle arbitrary value types, as long as they
can be encoded using Go's built-in `gob` package. The library automatically
handles serialization on writes and de-serialization on reads.

```go
err := memc.Set(client, "my/key/name", &Person{Name: "Bob"})
```

##### Reading a value from memcached.

The `memc` package will automatically convert the value `[]byte` into the type
of your Go variable.

```go
value, err := memc.Get[T](client, "my/key/name")
```

##### Incrementing/Decrementing a counter in memcached.

The `memc` package provides `Increment` and `Decrement` for increasing or
decreasing a value by a given delta. Note that the value must already be stored,
and must be in the form of an ASCII string representation of a number. The delta
value must be positive.

```go
err := memc.Set(client, "/my/counter", "100")
```

Using `Increment` to increase the counter value by 1.

```go
v, err := memc.Increment("/my/counter", 1)
// v is now 101
```

Using `Decrement` to decrease the value by 5.

```go
v, err := memc.Decrement("/my/counter", 5)
// v is now 96
```

##### Sharding memcached instances.

The memcached can handle sharding writes and reads across multiple memcached
instances. It does this by hashing the key space and deterministically choosing
an assigned instance. To enable this behavior simply give the `Client` each
memcached instance address.

```go
client := memc.New(
  []string{
    "10.0.0.1:11211",
    "10.0.0.2:11211",
    "10.0.0.3:11211",
  },
)
```

##### Configuring default expiration.

The `Client` sets a default expiration time on each value. This expiration time
can be configured when creating the client.

```go
client := memc.New(
  // ...
  SetDefaultTTL(2 * time.Minute),
)
```

##### Configuration idle connection pool.

The `Client` maintains an idle connection pool for each memcached instance it
has connected to. The number of idle connections to maintain per instance can be
adjusted.

```go
client := memc.New(
  // ...
  SetIdleConnections(4),
)
```

##### Closing the client.

The `Client` can be closed so that idle connections are closed and no longer
consuming resources. In-flight requests will be closed once complete. Once
closed the client cannot be reused.

```go
_ = memc.Close()
```

### License

The `cattlecloud.net/go/memc` module is open source under the [BSD-3-Clause](LICENSE) license.
