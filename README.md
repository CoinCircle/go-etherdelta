# go-etherdelta

> [EtherDelta](https://etherdelta.com/) client for [Golang](https://golang.org/).

[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/miguelmota/go-etherdelta/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/miguelmota/go-etherdelta?)](https://goreportcard.com/report/github.com/miguelmota/go-etherdelta) [![GoDoc](https://godoc.org/github.com/miguelmota/go-etherdelta?status.svg)](https://godoc.org/github.com/miguelmota/go-etherdelta)

## Documentation

[https://godoc.org/github.com/miguelmota/go-etherdelta](https://godoc.org/github.com/miguelmota/go-etherdelta)

# Examples

Take a look at the [tests](./etherdelta_test.go).

## Config

By default, the Ethereum provider is to `wss://mainnet.infura.io/ws` but you can override it by setting the `ETH_PROVIDER_URI` environment variable.


```bash
export ETH_PROVIDER_URI='wss://mainnet.infura.io/ws'
```

## Test

```bash
go test -v ./...
```

## Resources

- [EtherDelta API](https://github.com/etherdelta/etherdelta.github.io/blob/master/docs/API.md)
- [EtherDelta Smart Contracts](https://github.com/etherdelta/etherdelta.github.io/blob/master/docs/SMART_CONTRACT.md)

## License

MIT
