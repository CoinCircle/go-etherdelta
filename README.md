# go-etherdelta

> [EtherDelta](https://etherdelta.com/) client for [Golang](https://golang.org/).

[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/miguelmota/go-etherdelta/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/miguelmota/go-etherdelta?)](https://goreportcard.com/report/github.com/miguelmota/go-etherdelta) [![GoDoc](https://godoc.org/github.com/miguelmota/go-etherdelta?status.svg)](https://godoc.org/github.com/miguelmota/go-etherdelta)

## Documentation

[https://godoc.org/github.com/miguelmota/go-etherdelta](https://godoc.org/github.com/miguelmota/go-etherdelta)

## Getting started

```go
package main

import (
	"log"

	ed "github.com/miguelmota/go-etherdelta"
)

func main() {
	service := ed.New(&ed.Options{
		ProviderURI: "wss://mainnet.infura.io/ws",
	})

	orders, err := service.GetOrderBook(&ed.GetOrderBookOpts{
		TokenAddress: "0x0d8775f648430679a709e98d2b0cb6250d2887ef",
	})

	if err != nil {
		panic(err)
	}

	log.Println(orders)
}
```

## ForkDelta

A [ForkDelta](https://forkdelta.github.io/) client is avaiable:

```go
service := ed.NewForkDelta(&ed.Options{
  ProviderURI: "wss://mainnet.infura.io/ws",
})
```

## Examples

Take a look at the [tests](./etherdelta_test.go).

## Config

You can pass the `ProviderURI` property to the EtherDelta constructor options. If this is not set, then the service will read the `ETH_PROVIDER_URI` environment variable, otherwise the Ethereum provider is set to `wss://mainnet.infura.io/ws` by default.

## Test

```bash
go test -v ./...
```

## FAQ

- Q: Why do I get empty results sometimes?

    - A: Sometimes the EtherDelta websocket connection randomly disconnects. It can be unreliable at times.

- Q: It's completely not working anymore!

    - A: EtherDelta may have changed their API or websocket endpoint.

## Resources

- [EtherDelta API](https://github.com/etherdelta/etherdelta.github.io/blob/master/docs/API.md)
- [EtherDelta Smart Contracts](https://github.com/etherdelta/etherdelta.github.io/blob/master/docs/SMART_CONTRACT.md)

## License

MIT
