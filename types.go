package etherdelta

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/graarh/golang-socketio"
	"github.com/shopspring/decimal"
)

// OrderBook Order Book struct
type OrderBook struct {
	Buys  []Order `json:"buys"`
	Sells []Order `json:"sells"`
}

// Order Individual order struct
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

// OrderPost Order struct for posting a new order to EtherDelta
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

// TokenTicker Ticker infor for token
type TokenTicker struct {
	Ask           *decimal.Decimal `json:"ask"`
	BaseVolume    *decimal.Decimal `json:"baseVolume"`
	Bid           *decimal.Decimal `json:"bid"`
	Last          *decimal.Decimal `json:"last"`
	PercentChange *decimal.Decimal `json:"percentChange"`
	QuoteVolume   *decimal.Decimal `json:"quoteVolume"`
	TokenAddress  string           `json:"tokenAddr"`
}

// GetOrderBookOpts Options for getting order book
type GetOrderBookOpts struct {
	TokenAddress string
	UserAddress  string
}

// GetTokenTickerOpts Options for getting token ticker
type GetTokenTickerOpts struct {
	TokenSymbol string
}

// GetTokenPriceOpts Options for getting token price
type GetTokenPriceOpts struct {
	TokenSymbol string
}

// GetTokenBalanceOpts Options for getting token balance
type GetTokenBalanceOpts struct {
	TokenAddress string
	UserAddress  string
}

// PostOrderOpts Options for posting an order
type PostOrderOpts struct {
	Order        *OrderPost
	TokenAddress string
	UserAddress  string
}

// MakeOrderOpts Options for making an order
type MakeOrderOpts struct {
	PrivateKey   string
	TokenAddress string
	UserAddress  string
	Amount       *decimal.Decimal
	EthCost      *decimal.Decimal
}

// CancelOrderOpts Options for cancelling an order
type CancelOrderOpts struct {
	PrivateKey string
	Order      *OrderPost
}

// MakeTrade Options for making a trade
type MakeTradeOpts struct {
	Auth    *bind.TransactOpts
	Order   *OrderPost
	EthCost *big.Int
}

// DepositEthOpts Options for depositing ETH
type DepositEthOpts struct {
	Auth *bind.TransactOpts
}

// WithdrawTokenOpts Options for withdrawing token
type WithdrawTokenOpts struct {
	Auth         *bind.TransactOpts
	TokenAddress string
	TokenAmount  *big.Int
}

type wsClient struct {
	client *gosocketio.Client
}

type wsRequest struct {
	EmitTopic   string      `json:"emitTopic"`
	EmitBody    *wsEmitBody `json:"emitBody"`
	ListenTopic string      `json:"listenTopic"`
}

type wsEmitBody struct {
	Token string     `json:"token"`
	User  string     `json:"user"`
	Order *OrderPost `json:"-"` // omit
}

type wsMessage interface{}

type wsResponse struct {
	Message wsMessage
	Error   error
}
