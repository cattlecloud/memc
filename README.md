# memc

`memc` is a modern and generics enabled memcached client library for Go.

_requires go1.23+_

October 2024:
(!) This package is very new and may contain bugs and missing features.

### Getting Started

The `memc` package can be added to a project by running:

```shell
go get noxide.lol/go/memc@latest
```

### Examples

##### Setting a value in memcached.

```go
client := memc.New(
  SetServer("localhost:11211"),
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

##### Sharding memcached instances.

The memcached can handle sharding writes and reads across multiple memcached
instances. It does this by hashing the key space and deterministically choosing
an assigned instance. To enable this behavior simply give the `Client` each
memcached instance address.

```go
client := memc.New(
  SetServer("10.0.0.101:11211"),
  SetServer("10.0.0.102:11211"),
  SetServer("10.0.0.103:11211"),
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

The `noxide.lol/go/memc` module is open source under the [BSD-3-Clause](LICENSE) license.
