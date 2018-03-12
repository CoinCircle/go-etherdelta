package etherdelta

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	contracts "github.com/miguelmota/go-etherdelta/contracts"
	"github.com/miguelmota/go-etherdelta/helpers"
	"github.com/miguelmota/go-etherdelta/utils"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
	"github.com/shopspring/decimal"
)

const etherDeltaContractAddress = "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819"

var httpClient = &http.Client{Timeout: 10 * time.Second}
var maxTries = 3
var cache *freecache.Cache

// Service service struct
type Service struct {
	client       *ethclient.Client
	instance     *contracts.EtherDelta
	websocketURI string
}

// Options service options
type Options struct {
	ProviderURI  string
	websocketURI string
}

const etherDeltaWSURL = "wss://socket.etherdelta.com/socket.io/?EIO=3&transport=websocket"
const forkDeltaWSURL = "wss://api.forkdelta.com/socket.io/?EIO=3&transport=websocket"

// newService returns a new service
func newService(opts *Options) *Service {
	if opts.ProviderURI != "" {
		helpers.SetClientProviderURI(opts.ProviderURI)
	}
	client := helpers.Client
	instance, err := contracts.NewEtherDelta(common.HexToAddress(etherDeltaContractAddress), client)

	if err != nil {
		log.Fatalf("Could not initialize EtherDelta contract, got error %s", err)
	}

	cache = freecache.NewCache(0)
	return &Service{
		client:       client,
		instance:     instance,
		websocketURI: opts.websocketURI,
	}
}

// New returns new service
func New(opts *Options) *Service {
	opts.websocketURI = etherDeltaWSURL
	return newService(opts)
}

// NewForkDelta returns new ForkDelta service
func NewForkDelta(opts *Options) *Service {
	opts.websocketURI = forkDeltaWSURL
	return newService(opts)
}

// GetOrderBook Get the Order Book
func (s *Service) GetOrderBook(opts *GetOrderBookOpts) (*OrderBook, error) {
	log.Debug("Attempting websocket connection to get order book")

	orderBook := OrderBook{
		Buys:  []Order{},
		Sells: []Order{},
	}

	var target map[string]interface{}
	tries := 0

	log.Debug("Fetching EtherDelta orderbook for token %s", opts.TokenAddress)

	// retry if error or null response
	for tries < maxTries {
		if tries > 0 {
			log.Warnf("Try #%v", tries+1)
		}

		wsrequest := &wsRequest{
			EmitTopic: "getMarket",
			EmitBody: &wsEmitBody{
				Token: opts.TokenAddress,
				User:  opts.UserAddress,
			},
			ListenTopic: "market",
		}

		result, err := s.makeWSRequest(wsrequest)

		if err != nil {
			tries = tries + 1
			continue
		}

		jsonStr, err := json.Marshal(result)

		if err != nil {
			tries = tries + 1
			continue
		}

		reader := bytes.NewReader(jsonStr)
		json.NewDecoder(reader).Decode(&target)

		if target == nil {
			tries = tries + 1
			continue
		}

		break
	}

	if target == nil {
		return &orderBook, errors.New("Error getting orderbook, got null")
	}

	orderTypes := []string{"sells", "buys"}

	for _, orderType := range orderTypes {
		if target["orders"] == nil {
			return &orderBook, errors.New("Error getting orders, got null")
		}

		allOrders := target["orders"].(map[string]interface{})
		ordersTypeArray := allOrders[orderType].([]interface{})

		for i := range ordersTypeArray {
			orderJSON := ordersTypeArray[i].(map[string]interface{})
			order := Order{}

			// TODO: improve this

			id, ok := orderJSON["id"].(string)

			if ok {
				order.Id = id
			}

			amount, ok := orderJSON["amount"].(string)

			if ok {
				order.Amount = amount
			}

			price, ok := orderJSON["price"].(string)

			if ok {
				order.Price = price
			}

			tokenGet, ok := orderJSON["tokenGet"].(string)

			if ok {
				order.TokenGet = tokenGet
			}

			amountGet, ok := orderJSON["amountGet"].(string)

			if ok {
				order.AmountGet = amountGet
			}

			tokenGive, ok := orderJSON["tokenGive"].(string)

			if ok {
				order.TokenGive = tokenGive
			}

			amountGive, ok := orderJSON["amountGive"].(string)

			if ok {
				order.AmountGive = amountGive
			}

			expires, ok := orderJSON["expires"].(string)

			if ok {
				order.Expires = expires
			}

			nonce, ok := orderJSON["nonce"].(string)

			if ok {
				order.Nonce = nonce
			}

			v := int(orderJSON["v"].(float64))
			order.V = v

			r, ok := orderJSON["r"].(string)

			if ok {
				order.R = r
			}

			s, ok := orderJSON["s"].(string)
			if ok {
				order.S = s
			}

			user, ok := orderJSON["user"].(string)

			if ok {
				order.User = user
			}

			updated, ok := orderJSON["updated"].(string)

			if ok {
				order.Updated = updated
			}

			availableVolume, ok := orderJSON["availableVolume"].(string)

			if ok {
				order.AvailableVolume = availableVolume
			}

			ethAvailableVolume, ok := orderJSON["ethAvailableVolume"].(string)

			if ok {
				order.EthAvailableVolume = ethAvailableVolume
			}

			availableVolumeBase, ok := orderJSON["availableVolumeBase"].(string)

			if ok {
				order.AvailableVolumeBase = availableVolumeBase
			}

			amountFilled, ok := orderJSON["amountFilled"].(string)

			if ok {
				order.AmountFilled = amountFilled
			}

			if orderType == "sells" {
				orderBook.Sells = append(orderBook.Sells, order)
			}

			if orderType == "buys" {
				orderBook.Buys = append(orderBook.Buys, order)
			}
		}
	}

	return &orderBook, nil
}

// GetTokenTicker Get ticker info for token
func (s *Service) GetTokenTicker(opts *GetTokenTickerOpts) (*TokenTicker, error) {
	log.Debug("Attempting websocket connection to get token ticker")

	tokenTicker := &TokenTicker{}

	var target map[string]interface{}
	tries := 0

	log.Debugf("Fetching EtherDelta ticker for token %s", opts.TokenSymbol)

	// retry if error or null response
	for tries < maxTries {
		if tries > 0 {
			log.Warnf("Try #%v", tries+1)
		}

		wsrequest := &wsRequest{
			EmitTopic: "getMarket",
			EmitBody: &wsEmitBody{
				Token: "",
				User:  "",
			},
			ListenTopic: "market",
		}

		result, err := s.makeWSRequest(wsrequest)

		if err != nil {
			tries = tries + 1
			continue
		}

		jsonStr, err := json.Marshal(result)

		if err != nil {
			tries = tries + 1
			continue
		}

		reader := bytes.NewReader(jsonStr)
		json.NewDecoder(reader).Decode(&target)

		if target == nil {
			tries = tries + 1
			continue
		}

		if target == nil {
			tries = tries + 1
			continue
		}

		if target["returnTicker"] == nil {
			tries = tries + 1
			continue
		}

		break
	}

	if target == nil {
		return tokenTicker, errors.New("Error getting token ticker, got null")
	}

	if target["returnTicker"] == nil {
		return tokenTicker, errors.New("Error getting ticker property, got null")
	}

	tickers := target["returnTicker"].(map[string]interface{})
	ticker := tickers[fmt.Sprintf("ETH_%s", opts.TokenSymbol)].(map[string]interface{})

	if utils.KeyExists(ticker, "ask") {
		if f, ok := ticker["ask"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.Ask = &v
		}
		if f, ok := ticker["ask"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.Ask = &v
			}
		}
	}

	if utils.KeyExists(ticker, "baseVolume") {
		if f, ok := ticker["baseVolume"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.BaseVolume = &v
		}
		if f, ok := ticker["baseVolume"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.BaseVolume = &v
			}
		}
	}

	if utils.KeyExists(ticker, "bid") {
		if f, ok := ticker["bid"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.Bid = &v
		}
		if f, ok := ticker["bid"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.Bid = &v
			}
		}
	}

	if utils.KeyExists(ticker, "last") {
		if f, ok := ticker["last"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.Last = &v
		}
		if f, ok := ticker["last"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.Last = &v
			}
		}
	}

	if utils.KeyExists(ticker, "percentChange") {
		if f, ok := ticker["percentChange"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.PercentChange = &v
		}
		if f, ok := ticker["percentChange"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.PercentChange = &v
			}
		}
	}

	if utils.KeyExists(ticker, "percentChange") {
		if f, ok := ticker["quoteVolume"].(float64); ok {
			v := decimal.NewFromFloat(f)
			tokenTicker.QuoteVolume = &v
		}
		if f, ok := ticker["quoteVolume"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				v := decimal.NewFromFloat(parsed)
				tokenTicker.QuoteVolume = &v
			}
		}
	}

	if utils.KeyExists(ticker, "tokenAddr") {
		if s, ok := ticker["tokenAddr"].(string); ok {
			tokenTicker.TokenAddress = s
		}
	}

	return tokenTicker, nil
}

// GetTokenPrice Get last trade price for token
func (s *Service) GetTokenPrice(opts *GetTokenPriceOpts) (*decimal.Decimal, error) {
	var price decimal.Decimal

	cachekey := []byte(opts.TokenSymbol)
	cached, err := cache.Get(cachekey)

	if err == nil {
		price, err = decimal.NewFromString(string(cached))

		if err == nil {
			log.Warnf("Returning cached price for %s: %s", opts.TokenSymbol, price)
			return &price, nil
		}
	}

	getTokenTickerOpts := &GetTokenTickerOpts{
		TokenSymbol: "UKG",
	}

	ticker, err := s.GetTokenTicker(getTokenTickerOpts)

	if err != nil {
		return &price, err
	}

	price = *ticker.Last

	cache.Set(cachekey, []byte(price.String()), 10)

	return &price, nil
}

// GetTokenBalance Get token balance on EtherDelta for account
func (s *Service) GetTokenBalance(opts *GetTokenBalanceOpts) (*big.Int, error) {
	var (
		err               error
		etherDeltaBalance *big.Int
	)

	etherDeltaBalance, err = s.instance.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress(opts.TokenAddress), common.HexToAddress(opts.UserAddress))

	if err != nil {
		return etherDeltaBalance, err
	}

	return etherDeltaBalance, nil
}

// PostOrder Post an order to EtherDelta
func (s *Service) PostOrder(opts *PostOrderOpts) (string, error) {
	wsrequest := &wsRequest{
		EmitTopic: "message",
		EmitBody: &wsEmitBody{
			Token: opts.TokenAddress,
			User:  opts.UserAddress,
			Order: opts.Order,
		},
		ListenTopic: "messageResult",
	}

	postResponse, err := s.makeWSRequest(wsrequest)
	result := ""

	if err != nil {
		return result, err
	}

	if postResponse == nil {
		return result, errors.New("Got null response from post order")
	}

	result, ok := postResponse.(string)

	if !ok {
		return result, errors.New("Could not parse response into string")
	}

	return result, nil
}

// MakeOrder Generate an order and post it to EtherDelta
func (s *Service) MakeOrder(opts *MakeOrderOpts) (string, error) {
	var result string
	decimals, err := helpers.GetTokenDecimals(opts.TokenAddress)

	if err != nil {
		return result, err
	}

	latestBlockNumber, err := helpers.GetLatestBlockNumber()

	if err != nil {
		return result, err
	}

	ethAmountInWei := helpers.DecimalsToWei(opts.EthCost, big.NewInt(18)).String()
	ethAddress := "0x0000000000000000000000000000000000000000" // 0 address for eth
	amountWantInWei := helpers.DecimalsToWei(opts.Amount, decimals).String()
	expiresBlockNumber := int(new(big.Int).Add(latestBlockNumber, big.NewInt(20000)).Int64())
	nonce := utils.RandInt(1000000000, 10000000000)

	key, err := crypto.HexToECDSA(opts.PrivateKey)

	if err != nil {
		return result, err
	}

	orderPost := &OrderPost{
		AmountGive:      ethAmountInWei,
		TokenGive:       ethAddress,
		AmountGet:       amountWantInWei,
		TokenGet:        opts.TokenAddress,
		ContractAddress: etherDeltaContractAddress,
		Expires:         expiresBlockNumber,
		Nonce:           nonce,
		User:            opts.UserAddress,
		V:               0,
		R:               "",
		S:               "",
	}

	data := solsha3.ConcatByteSlices(
		solsha3.Address(orderPost.ContractAddress),
		solsha3.Address(orderPost.TokenGet),
		solsha3.Uint256FromString(orderPost.AmountGet),
		solsha3.Address(orderPost.TokenGive),
		solsha3.Uint256FromString(orderPost.AmountGive),
		solsha3.Uint256(big.NewInt(int64(orderPost.Expires))),
		solsha3.Uint256(big.NewInt(int64(orderPost.Nonce))),
	)

	hash := sha256.Sum256(data)

	sha3HashWithPrefix := solsha3.SoliditySHA3(
		solsha3.ConcatByteSlices(
			// prefix is required
			[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n32")),
			hash[:],
		),
	)

	msg := sha3HashWithPrefix

	sig, err := crypto.Sign(msg, key)

	if err != nil {
		return result, fmt.Errorf("Sign error: %s", err)
	}

	rsv := helpers.GetSigRSV(sig)

	orderPost.R = rsv.R
	orderPost.S = rsv.S
	orderPost.V = rsv.V

	recoveredPub, err := crypto.Ecrecover(msg, sig)

	if err != nil {
		return result, fmt.Errorf("ECRecover error: %s", err)
	}

	pubKey := crypto.ToECDSAPub(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	addr := common.HexToAddress(opts.UserAddress)

	if addr != recoveredAddr {
		return result, fmt.Errorf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}

	postOrderOpts := &PostOrderOpts{
		Order:        orderPost,
		TokenAddress: opts.TokenAddress,
		UserAddress:  opts.UserAddress,
	}

	result, err = s.PostOrder(postOrderOpts)

	if err != nil {
		return result, err
	}

	log.Infof("Post order message response: %s", result)

	return result, nil
}

// CancelOrder Cancel an order on EtherDelta
func (s *Service) CancelOrder(opts *CancelOrderOpts) ([]byte, error) {
	var txHash []byte
	order := opts.Order
	key, err := crypto.HexToECDSA(opts.PrivateKey)

	if err != nil {
		return txHash, err
	}

	auth := *bind.NewKeyedTransactor(key)

	amountGet := new(big.Int)
	amountGet, _ = amountGet.SetString(order.AmountGet, 10)

	amountGive := new(big.Int)
	amountGive, _ = amountGive.SetString(order.AmountGive, 10)

	expires := big.NewInt(int64(order.Expires))
	nonce := big.NewInt(int64(order.Nonce))

	V := uint8(order.V)
	R := [32]byte{}
	S := [32]byte{}

	copy(R[:], common.FromHex(order.R)[:])
	copy(S[:], common.FromHex(order.S)[:])

	tx, err := s.instance.CancelOrder(
		&auth,
		common.HexToAddress(order.TokenGet),
		amountGet,
		common.HexToAddress(order.TokenGive),
		amountGive,
		expires,
		nonce,
		V,
		R,
		S,
	)

	log.Infof("CancelOrder TX: %s", tx)

	txHash = tx.Hash().Bytes()

	return txHash, err
}

// MakeTrade Make an order trade on EtherDelta
func (s *Service) MakeTrade(opts *MakeTradeOpts) ([]byte, error) {
	var txHash []byte
	order := opts.Order

	amountGet := new(big.Int)
	amountGet, _ = amountGet.SetString(order.AmountGet, 10)

	amountGive := new(big.Int)
	amountGive, _ = amountGive.SetString(order.AmountGive, 10)

	expires := big.NewInt(int64(order.Expires))
	nonce := big.NewInt(int64(order.Nonce))

	V := uint8(order.V)
	R := [32]byte{}
	S := [32]byte{}

	copy(R[:], common.FromHex(order.R)[:])
	copy(S[:], common.FromHex(order.S)[:])

	user := common.HexToAddress(order.User)
	sender := common.HexToAddress(order.User)

	tokenGet := common.HexToAddress(order.TokenGet)
	tokenGive := common.HexToAddress(order.TokenGive)

	isValid, err := s.instance.TestTrade(
		&bind.CallOpts{},
		tokenGet,
		amountGet,
		tokenGive,
		amountGive,
		expires,
		nonce,
		user,
		V,
		R,
		S,
		big.NewInt(0),
		sender,
	)

	if !isValid {
		return txHash, errors.New("Trade is not valid")
	}

	tx, err := s.instance.Trade(
		opts.Auth,
		tokenGet,
		amountGet,
		tokenGive,
		amountGive,
		expires,
		nonce,
		user,
		V,
		R,
		S,
		opts.EthCost, // in wei
	)

	if err != nil {
		return txHash, fmt.Errorf("Trade failed, got error: %s", err)
	}

	log.Infof("MakeTrade TX: %s", tx)

	txHash = tx.Hash().Bytes()

	return txHash, err
}

// DepositEth Deposit ETH to EtherDelta
func (s *Service) DepositEth(opts *DepositEthOpts) ([]byte, error) {
	var txHash []byte

	tx, err := s.instance.Deposit(opts.Auth)

	if err != nil {
		return txHash, err
	}

	txHash = tx.Hash().Bytes()

	log.Infof("Deposited to EtherDelta, TX: %s", tx)

	return txHash, nil
}

// WithdrawToken Withdraw token from EtherDelta
func (s *Service) WithdrawToken(opts *WithdrawTokenOpts) ([]byte, error) {
	var txHash []byte
	tx, err := s.instance.WithdrawToken(
		opts.Auth,
		common.HexToAddress(opts.TokenAddress),
		opts.TokenAmount,
	)

	if err != nil {
		return txHash, err
	}

	txHash = tx.Hash().Bytes()

	log.Infof("Withdrew tokens from EtherDelta, TX: %s", tx)

	return txHash, nil
}

// GetJSON Make a JSON request
func (s *Service) GetJSON(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/534.24 (KHTML, like Gecko) Chrome/11.0.696.0 Safari/534.24")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString, err
}

// ParseStringExpNotation Parse string in exponential notation
func (s *Service) ParseStringExpNotation(str string) string {
	// etherdelta will return a string such as "2.8e+21" if the number is too big
	{
		re := regexp.MustCompile(`e+`)
		match := re.Split(str, 2)

		if len(match) > 1 {
			n, _ := decimal.NewFromString(match[0])
			exp, _ := decimal.NewFromString(match[1])
			n = n.Mul(decimal.NewFromFloat(10).Pow(exp))

			str = n.String()
		}
	}

	{
		re := regexp.MustCompile("[0-9]-")
		match := re.Split(str, 2)

		if len(match) == 1 {
			return str
		}

		re = regexp.MustCompile("-0")
		match = re.Split(str, 2)

		if len(match) > 1 {
			n, _ := decimal.NewFromString(match[0])
			exp, _ := decimal.NewFromString(match[1])
			n = n.Mul(decimal.NewFromFloat(0.1).Pow(exp))

			str = n.String()
		}
	}

	return str
}

func (s *Service) getTickerAPI() (string, error) {
	url := "https://api.etherdelta.com/returnTicker"
	jsonStr, err := s.GetJSON(url)

	return jsonStr, err
}

func (s *Service) makeWSRequest(wsrequest *wsRequest) (interface{}, error) {
	isConnected := make(chan bool, 1)
	client := newWSClient(s.websocketURI, isConnected)

	switch <-isConnected {
	case false:
		return nil, errors.New("Could not establish connection")
	default:
	}

	message := make(chan *wsResponse)

	client.EmitListenOnceAndClose(wsrequest.ListenTopic, wsrequest.EmitBody, message, wsrequest.EmitTopic)
	result := <-message

	if result.Error != nil {
		return nil, result.Error
	}

	if result.Message == nil {
		return nil, errors.New("Empty response from EtherDelta")
	}

	return result.Message, nil
}
