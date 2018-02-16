package main

import (
	"log"

	etherdelta "github.com/miguelmota/go-etherdelta"
)

func main() {
	service := etherdelta.New(&etherdelta.Options{
		ProviderURI: "wss://mainnet.infura.io/ws",
	})

	orders, err := service.GetOrderBook(&etherdelta.GetOrderBookOpts{
		TokenAddress: "0x0d8775f648430679a709e98d2b0cb6250d2887ef",
	})

	if err != nil {
		panic(err)
	}

	log.Println(orders)
}
