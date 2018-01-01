# go-etherdelta

> [EtherDelta](https://etherdelta.com/) client for [Golang](https://golang.org/).

**NOTICE: currency in pre-alpha stage. Not ready for production.**

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
