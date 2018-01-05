package etherdelta

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/graarh/golang-socketio"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	Buys  []Order `json:"buys"`
	Sells []Order `json:"sells"`
}

type Order struct {
	Id                  string `json:"id"`
	Amount              string `json:"amount"`
	Price               string `json:"price"`
	TokenGet            string `json:"tokenGet"`
	AmountGet           string `json:"amountGet"`
	TokenGive           string `json:"tokenGive"`
	AmountGive          string `json:"amountGive"`
	Expires             string `json:"expires"`
	Nonce               string `json:"nonce"`
	V                   int    `json:"v"`
	R                   string `json:"r"`
	S                   string `json:"s"`
	User                string `json:"user"`
	Updated             string `json:"updated"`
	AvailableVolume     string `json:"availableVolume"`
	EthAvailableVolume  string `json:"ethAvailableVolume"`
	AvailableVolumeBase string `json:"availableVolumeBase"`
	AmountFilled        string `json:"amountFilled"`
}

type OrderPost struct {
	AmountGet       string `json:"amountGet"`
	AmountGive      string `json:"amountGive"`
	TokenGet        string `json:"tokenGet"`
	TokenGive       string `json:"tokenGive"`
	ContractAddress string `json:"contractAddr"`
	Expires         int    `json:"expires"`
	Nonce           int    `json:"nonce"`
	User            string `json:"user"`
	V               int    `json:"v"`
	R               string `json:"r"`
	S               string `json:"s"`
}

type TokenTicker struct {
	Ask           *decimal.Decimal `json:"ask"`
	BaseVolume    *decimal.Decimal `json:"baseVolume"`
	Bid           *decimal.Decimal `json:"bid"`
	Last          *decimal.Decimal `json:"last"`
	PercentChange *decimal.Decimal `json:"percentChange"`
	QuoteVolume   *decimal.Decimal `json:"quoteVolume"`
	TokenAddress  string           `json:"tokenAddr"`
}

type GetOrderBookOpts struct {
	TokenAddress string
	UserAddress  string
}

type GetTokenTickerOpts struct {
	TokenSymbol string
}

type GetTokenPriceOpts struct {
	TokenSymbol string
}

type GetTokenBalanceOpts struct {
	TokenAddress string
	UserAddress  string
}

type PostOrderOpts struct {
	Order        *OrderPost
	TokenAddress string
	UserAddress  string
}

type MakeOrderOpts struct {
	PrivateKey   string
	TokenAddress string
	UserAddress  string
	Amount       *decimal.Decimal
	EthCost      *decimal.Decimal
}

type CancelOrderOpts struct {
	PrivateKey string
	Order      *OrderPost
}

type MakeTradeOpts struct {
	Auth    *bind.TransactOpts
	Order   *OrderPost
	EthCost *big.Int
}

type DepositEthOpts struct {
	Auth *bind.TransactOpts
}

type WithdrawTokenOpts struct {
	Auth         *bind.TransactOpts
	TokenAddress string
	TokenAmount  *big.Int
}

type WSClient struct {
	client *gosocketio.Client
}

type WSRequest struct {
	EmitTopic   string      `json:"emitTopic"`
	EmitBody    *WSEmitBody `json:"emitBody"`
	ListenTopic string      `json:"listenTopic"`
}

type WSEmitBody struct {
	Token string     `json:"token"`
	User  string     `json:"user"`
	Order *OrderPost `json:"-"` // omit
}

type Message interface{}

type WSResponse struct {
	Message Message
	Error   error
}
