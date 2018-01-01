package etherdelta

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	contracts "github.com/miguelmota/go-etherdelta/contracts"
	"github.com/miguelmota/go-etherdelta/ethereum"
	"github.com/miguelmota/go-etherdelta/utils"
	"github.com/miguelmota/go-solidity-sha3"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var EDInstance *contracts.EtherDelta
var etherDeltaContractAddress string
var httpClient = &http.Client{Timeout: 10 * time.Second}
var maxTries = 3
var cache *freecache.Cache

func MakeWSRequest(wsrequest *WSRequest) (interface{}, error) {
	isConnected := make(chan bool, 1)
	client := NewWSClient(isConnected)

	switch <-isConnected {
	case false:
		return nil, errors.New("Could not establish connection")
	default:
	}

	message := make(chan *WSResponse)

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

func GetOrderBook(tokenAddress string, userAddress string) (*OrderBook, error) {
	log.Printf("Attempting websocket connection to get order book")

	orderBook := OrderBook{
		Buys:  []Order{},
		Sells: []Order{},
	}

	var target map[string]interface{}
	tries := 0

	log.Printf("Fetching EtherDelta orderbook for token %s", tokenAddress)

	// retry if error or null response
	for tries < maxTries {
		if tries > 0 {
			log.Printf("Try #%v", tries+1)
		}

		wsrequest := &WSRequest{
			EmitTopic: "getMarket",
			EmitBody: &WSEmitBody{
				Token: tokenAddress,
				User:  userAddress,
			},
			ListenTopic: "market",
		}

		result, err := MakeWSRequest(wsrequest)

		if err != nil {
			tries = tries + 1
			continue
		}

		jsonStr, err := json.Marshal(result)
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

		for i, _ := range ordersTypeArray {
			orderJson := ordersTypeArray[i].(map[string]interface{})
			order := Order{}

			// TODO: improve this

			id, ok := orderJson["id"].(string)

			if ok {
				order.Id = id
			}

			amount, ok := orderJson["amount"].(string)

			if ok {
				order.Amount = amount
			}

			price, ok := orderJson["price"].(string)

			if ok {
				order.Price = price
			}

			tokenGet, ok := orderJson["tokenGet"].(string)

			if ok {
				order.TokenGet = tokenGet
			}

			amountGet, ok := orderJson["amountGet"].(string)

			if ok {
				order.AmountGet = amountGet
			}

			tokenGive, ok := orderJson["tokenGive"].(string)

			if ok {
				order.TokenGive = tokenGive
			}

			amountGive, ok := orderJson["amountGive"].(string)

			if ok {
				order.AmountGive = amountGive
			}

			expires, ok := orderJson["expires"].(string)

			if ok {
				order.Expires = expires
			}

			nonce, ok := orderJson["nonce"].(string)

			if ok {
				order.Nonce = nonce
			}

			v := int(orderJson["v"].(float64))
			order.V = v

			r, ok := orderJson["r"].(string)

			if ok {
				order.R = r
			}

			s, ok := orderJson["s"].(string)
			if ok {
				order.S = s
			}

			user, ok := orderJson["user"].(string)

			if ok {
				order.User = user
			}

			updated, ok := orderJson["updated"].(string)

			if ok {
				order.Updated = updated
			}

			availableVolume, ok := orderJson["availableVolume"].(string)

			if ok {
				order.AvailableVolume = availableVolume
			}

			ethAvailableVolume, ok := orderJson["ethAvailableVolume"].(string)

			if ok {
				order.EthAvailableVolume = ethAvailableVolume
			}

			availableVolumeBase, ok := orderJson["availableVolumeBase"].(string)

			if ok {
				order.AvailableVolumeBase = availableVolumeBase
			}

			amountFilled, ok := orderJson["amountFilled"].(string)

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

func MakeOrder(tokenAddress string, amountWant *decimal.Decimal, userAddress string, ethCost *decimal.Decimal, privateKey string) error {
	decimals, err := eth.GetTokenDecimals(tokenAddress)

	if err != nil {
		return err
	}

	latestBlockNumber, err := eth.GetLatestBlockNumber()

	if err != nil {
		return err
	}

	ethAmountInWei := eth.DecimalsToWei(ethCost, big.NewInt(18)).String()
	ethAddress := "0x0000000000000000000000000000000000000000" // 0 address for eth
	amountWantInWei := eth.DecimalsToWei(amountWant, decimals).String()
	expiresBlockNumber := int(new(big.Int).Add(latestBlockNumber, big.NewInt(20000)).Int64())
	nonce := utils.RandInt(1000000000, 10000000000)

	key, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		return err
	}

	orderPost := &OrderPost{
		AmountGive:      ethAmountInWei,
		TokenGive:       ethAddress,
		AmountGet:       amountWantInWei,
		TokenGet:        tokenAddress,
		ContractAddress: etherDeltaContractAddress,
		Expires:         expiresBlockNumber,
		Nonce:           nonce,
		User:            userAddress,
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
		return errors.New(fmt.Sprintf("Sign error: %s", err))
	}

	rsv := eth.GetSigRSV(sig)

	orderPost.R = rsv.R
	orderPost.S = rsv.S
	orderPost.V = rsv.V

	recoveredPub, err := crypto.Ecrecover(msg, sig)

	if err != nil {
		return errors.New(fmt.Sprintf("ECRecover error: %s", err))
	}

	pubKey := crypto.ToECDSAPub(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	addr := common.HexToAddress(userAddress)

	if addr != recoveredAddr {
		return errors.New(fmt.Sprintf("Address mismatch: want: %x have: %x", addr, recoveredAddr))
	}

	postResponse, err := PostOrder(orderPost, tokenAddress, userAddress)

	if err != nil {
		return err
	}

	log.Printf("Post order message response: %s\n", postResponse)

	return nil
}

func PostOrder(payload *OrderPost, tokenAddress string, userAddress string) (string, error) {
	wsrequest := &WSRequest{
		EmitTopic: "message",
		EmitBody: &WSEmitBody{
			Token: tokenAddress,
			User:  userAddress,
			Order: *payload,
		},
		ListenTopic: "messageResult",
	}

	postResponse, err := MakeWSRequest(wsrequest)
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

func CancelOrder(order *OrderPost, privateKey string) error {
	key, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		return err
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

	tx, err := EDInstance.CancelOrder(
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

	log.Println(tx)

	return err
}

func MakeTrade(order *OrderPost, ethCostInWei *big.Int, privateKey string) error {
	key, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		return err
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

	user := common.HexToAddress(order.User)
	sender := common.HexToAddress(order.User)

	tokenGet := common.HexToAddress(order.TokenGet)
	tokenGive := common.HexToAddress(order.TokenGive)

	isValid, err := EDInstance.TestTrade(
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
		return errors.New("Trade is not valid")
	}

	auth.GasLimit = big.NewInt(210000)      // units
	auth.GasPrice = big.NewInt(35000000000) // wei
	auth.Value = big.NewInt(0)

	tx, err := EDInstance.Trade(
		&auth,
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
		ethCostInWei,
	)

	if err != nil {
		return errors.New(fmt.Sprintf("Trade failed, got error: %s", err))
	}

	log.Printf("Made trade, got tx %s\n", tx)

	return err
}

func GetTokenBalance(tokenAddress string, userAddress string) (*big.Int, error) {
	var (
		err               error
		etherDeltaBalance *big.Int
	)

	etherDeltaBalance, err = EDInstance.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress(tokenAddress), common.HexToAddress(userAddress))

	if err != nil {
		return etherDeltaBalance, err
	}

	return etherDeltaBalance, nil
}

func GetTokenTicker(tokenSymbol string) (*TokenTicker, error) {
	log.Printf("Attempting websocket connection to get token ticker")

	tokenTicker := &TokenTicker{}

	var target map[string]interface{}
	tries := 0

	log.Printf("Fetching EtherDelta ticker for token %s", tokenSymbol)

	// retry if error or null response
	for tries < maxTries {
		if tries > 0 {
			log.Printf("Try #%v", tries+1)
		}

		wsrequest := &WSRequest{
			EmitTopic: "getMarket",
			EmitBody: &WSEmitBody{
				Token: "",
				User:  "",
			},
			ListenTopic: "market",
		}

		result, err := MakeWSRequest(wsrequest)

		if err != nil {
			tries = tries + 1
			continue
		}

		jsonStr, err := json.Marshal(result)
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
	ticker := tickers[fmt.Sprintf("ETH_%s", tokenSymbol)].(map[string]interface{})

	if keyExists(ticker, "ask") {
		if f, ok := ticker["ask"].(float64); ok {
			tokenTicker.Ask = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["ask"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.Ask = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "baseVolume") {
		if f, ok := ticker["baseVolume"].(float64); ok {
			tokenTicker.BaseVolume = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["baseVolume"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.BaseVolume = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "bid") {
		if f, ok := ticker["bid"].(float64); ok {
			tokenTicker.Bid = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["bid"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.Bid = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "last") {
		if f, ok := ticker["last"].(float64); ok {
			tokenTicker.Last = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["last"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.Last = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "percentChange") {
		if f, ok := ticker["percentChange"].(float64); ok {
			tokenTicker.PercentChange = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["percentChange"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.PercentChange = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "percentChange") {
		if f, ok := ticker["quoteVolume"].(float64); ok {
			tokenTicker.QuoteVolume = decimal.NewFromFloat(f)
		}
		if f, ok := ticker["quoteVolume"].(string); ok {
			parsed, err := strconv.ParseFloat(f, 64)
			if err == nil {
				tokenTicker.QuoteVolume = decimal.NewFromFloat(parsed)
			}
		}
	}

	if keyExists(ticker, "tokenAddr") {
		if s, ok := ticker["tokenAddr"].(string); ok {
			tokenTicker.TokenAddress = s
		}
	}

	return tokenTicker, nil
}

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

func GetTokenPrice(tokenSymbol string) (decimal.Decimal, error) {
	var price decimal.Decimal

	cachekey := []byte(tokenSymbol)
	cached, err := cache.Get(cachekey)

	if err == nil {
		price, err = decimal.NewFromString(string(cached))

		if err == nil {
			log.Println("Returning cached price for %s: %s", tokenSymbol, price)
			return price, nil
		}
	}

	ticker, err := GetTokenTicker(tokenSymbol)

	if err != nil {
		return price, err
	}

	price = ticker.Last

	cache.Set(cachekey, []byte(price.String()), 10)

	return price, nil
}

func DepositEth(amountInEth decimal.Decimal, privateKey string) error {
	key, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		return err
	}

	auth := *bind.NewKeyedTransactor(key)

	auth.GasLimit = big.NewInt(210000)      // units
	auth.GasPrice = big.NewInt(35000000000) // wei
	auth.Value = eth.EthToWei(&amountInEth)

	tx, err := EDInstance.Deposit(&auth)

	log.Printf("Deposited to EtherDelta %s", tx)

	return nil
}

func WithdrawToken(tokenAddress string, tokenAmount *big.Int, privateKey string) error {
	key, err := crypto.HexToECDSA(privateKey)

	if err != nil {
		return err
	}

	auth := *bind.NewKeyedTransactor(key)

	auth.GasLimit = big.NewInt(210000)      // units
	auth.GasPrice = big.NewInt(35000000000) // wei
	auth.Value = big.NewInt(0)

	tx, err := EDInstance.WithdrawToken(
		&auth,
		common.HexToAddress(tokenAddress),
		tokenAmount,
	)

	log.Printf("Withdrew tokens from EtherDelta %s", tx)

	return nil
}

func GetJson(url string) (string, error) {
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

func GetTickerApi() error {
	url := "https://api.etherdelta.com/returnTicker"
	jsonStr, err := GetJson(url)

	log.Println(jsonStr)

	return err
}

func init() {
	etherDeltaContractAddress = "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819"
	instance, err := contracts.NewEtherDelta(common.HexToAddress(etherDeltaContractAddress), eth.Client)

	if err != nil {
		log.Fatalf("Could not initialize EtherDelta contract, got error %s", err)
	}

	EDInstance = instance

	cache = freecache.NewCache(1e10)
}

func ParseStringExpNotation(str string) string {
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
