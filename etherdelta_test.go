package etherdelta

import (
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/miguelmota/go-solidity-sha3"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"testing"
)

var userAddress = ""
var privateKey = ""

func TestGetTokenBalance(t *testing.T) {
	t.Skip("Skipping GetTokenBalance")

	// BAT token
	tokenAddress := "0x0d8775f648430679a709e98d2b0cb6250d2887ef"

	balance, err := GetTokenBalance(tokenAddress, userAddress)

	if err != nil {
		t.Errorf("Got error", err)
	}

	if balance.Cmp(big.NewInt(0)) != 1 {
		t.Error("Expected balance to be greater than 0")
	}
}

func TestGetOrderBook(t *testing.T) {
	//t.Skip("Skipping GetOrderBook")

	// BAT token
	tokenAddress := "0x0d8775f648430679a709e98d2b0cb6250d2887ef"

	orders, err := GetOrderBook(tokenAddress, userAddress)

	if err != nil {
		t.Error(err)
	}

	buysSize := len(orders.Buys)
	sellsSize := len(orders.Sells)

	if buysSize <= 0 {
		t.Errorf("Expected Buys size to be bigger than 0, got %s", buysSize)
	}

	if sellsSize <= 0 {
		t.Errorf("Expected Sells size to be bigger than 0, got %s", sellsSize)
	}

	//t.Log(orders)
}

func TestGetTokenPrice(t *testing.T) {
	//t.Skip("Skipping GetTokenPrice")
	tokenSymbol := "UKG"

	price, err := GetTokenPrice(tokenSymbol)

	if err != nil {
		t.Errorf("Got err:", err)
	}

	if price.LessThanOrEqual(decimal.NewFromFloat(0)) {
		t.Errorf("Expected price to be greater than 0, instead got %s", price)
	}
}

func TestGetTokenTicker(t *testing.T) {
	//t.Skip("Skipping GetTokenPrice")
	tokenSymbol := "UKG"
	ticker, err := GetTokenTicker(tokenSymbol)

	if err != nil {
		t.Errorf("Got err:", err)
	}

	if ticker.Last.LessThanOrEqual(decimal.NewFromFloat(0)) {
		t.Errorf("Expected ticker last amount to be greater than 0, instead got %s", ticker.Last)
	}
}

func TestMakerOder(t *testing.T) {
	// disabled. enable when needing to test for reals
	t.Skip("skipping MakeOrder")

	// BAT token
	tokenAddress := "0x0d8775f648430679a709e98d2b0cb6250d2887ef"

	amount := decimal.NewFromFloat(300)
	ethCost := decimal.NewFromFloat(300 * 0.0003) // total tokens * price per token

	err := MakeOrder(tokenAddress, &amount, userAddress, &ethCost, privateKey)

	if err != nil {
		t.Errorf("Got error", err)
	}
}

func TestCancelOrder(t *testing.T) {
	// disabled. enable when needing to test for reals
	t.Skip("Skipping CancelOrder")

	order := &OrderPost{
		AmountGet:       "400000000000000000000",
		AmountGive:      "120000000000000000",
		TokenGet:        "0x0d8775f648430679a709e98d2b0cb6250d2887ef",
		TokenGive:       "0x0000000000000000000000000000000000000000",
		ContractAddress: "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819",
		Expires:         4729364,
		Nonce:           3747305518,
		User:            "0x5dbaD7F0D53934887C5D48ADE7a7bD5D7292a265",
		V:               27,
		R:               "0xd4c91f068099e1a39f458c02a1d30e4a0d4b7f4cc7b2097d09cc94d2c3afab30",
		S:               "0x6662a815783a93b24cdad22d8a55fa3d9f4bce496fb1d66644fdbaa9a3ae9060",
	}

	err := CancelOrder(order, privateKey)

	if err != nil {
		t.Error(err)
	}
}

func TestMakeTrade(t *testing.T) {
	// disabled. enable when needing to test for reals
	t.Skip("Skipping MakeTrade")

	// BAT token
	tokenAddress := "0x0d8775f648430679a709e98d2b0cb6250d2887ef"
	//tokenAddress := ""

	userAddress := "" // get all orders

	orders, err := GetOrderBook(tokenAddress, userAddress)

	if err != nil {
		t.Error(err)
	}

	// 0 index is best buy order
	order := orders.Sells[0]

	t.Log("Order to buy")
	t.Log(ParseStringExpNotation(order.AmountGet))
	t.Log(ParseStringExpNotation(order.AmountGive))
	t.Log(order.TokenGet)
	t.Log(order.TokenGive)
	t.Log(order.Expires)
	t.Log(order.User)
	t.Log(order.V)
	t.Log(order.R)
	t.Log(order.S)

	contractAddress := "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819"
	expires, _ := strconv.Atoi(order.Expires)
	nonce, _ := strconv.Atoi(order.Nonce)

	orderPost := &OrderPost{
		AmountGet:       ParseStringExpNotation(order.AmountGet),  // total cost in eth
		AmountGive:      ParseStringExpNotation(order.AmountGive), // how much we'll get
		TokenGet:        order.TokenGet,
		TokenGive:       order.TokenGive,
		ContractAddress: contractAddress,
		Expires:         expires,
		Nonce:           nonce,
		User:            order.User,
		V:               order.V,
		R:               order.R,
		S:               order.S,
	}

	tokenAmountInWei := big.NewInt(0)

	err = MakeTrade(orderPost, tokenAmountInWei, privateKey)

	if err != nil {
		t.Errorf("Trade failed, got error", err)
	}
}

func TestDepositEth(t *testing.T) {
	t.Skip("Skipping DepositEth")
	amount := decimal.NewFromFloat(0.02)

	err := DepositEth(amount, privateKey)

	if err != nil {
		t.Error(err)
	}
}

func TestWithdrawToken(t *testing.T) {
	// disabled. enable when needing to test for reals
	t.Skip("Skipping WithdrawToken")

	// UKG
	tokenAddress := "0x24692791bc444c5cd0b81e3cbcaba4b04acd1f3b"
	tokenAmount, err := GetTokenBalance(tokenAddress, userAddress)

	if err != nil {
		t.Errorf("Error getting token balance, got %s", err)
	}

	err = WithdrawToken(tokenAddress, tokenAmount, privateKey)

	if err != nil {
		t.Errorf("Error withrawing token, got %s", err)
	}
}

func TestSignature(t *testing.T) {
	// derived from js example:
	// https://gist.github.com/zackcoburn/c29cad5c18785d8308949cbcc26e4f23

	orderPost := &OrderPost{
		AmountGive:      "300000000000000000",
		TokenGive:       "0x0000000000000000000000000000000000000000",
		AmountGet:       "1000000000000000000",
		TokenGet:        "0x8f3470a7388c05ee4e7af3d01d8c722b0ff52374",
		ContractAddress: "0x8d12a197cb00d4747a1fe03395095ce2a5cc6819",
		Expires:         5000000,
		Nonce:           123456,
		User:            "0x88ebA77DC1b9E627e963020cE1f8dE9Fcf25edbD",

		// not yet set
		V: 0,
		R: "",
		S: "",
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

	expectedHash := "b60215f455a8b0e90a79df0f28933397059e070b23029820526165f817af6759"
	gotHash := common.Bytes2Hex(hash[:])

	if expectedHash != gotHash {
		t.Errorf("Expected %s, got %s", expectedHash, gotHash)
	}

	sha3HashNoPrefix := solsha3.SoliditySHA3(
		hash[:],
	)

	gotSha3HashNoPrefix := common.Bytes2Hex(sha3HashNoPrefix)
	expectedSha3HashNoPrefix := "d2a40424c4137f4269cf44d4f745f89baf5bc2a7df29083c99bf19e6ffa689f1"

	if expectedSha3HashNoPrefix != gotSha3HashNoPrefix {
		t.Errorf("Expected %s, got %s", expectedHash, gotHash)
	}

	combined := solsha3.ConcatByteSlices(
		[]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n32")),
		hash[:],
	)

	gotCombined := fmt.Sprintf("%x", combined)
	combinedExpected := "19457468657265756d205369676e6564204d6573736167653a0a3332b60215f455a8b0e90a79df0f28933397059e070b23029820526165f817af6759"

	if combinedExpected != gotCombined {
		t.Errorf("Expected %s, got %s", combinedExpected, gotCombined)
	}

	sha3HashWithPrefix := solsha3.SoliditySHA3(combined)

	sha3HashWithPrefixExpected := "757a9498e7b511f9c74a414de4207ee392642d4ef53510aba55e66330f85d3e8"
	gotSha3HashWithPrefix := common.Bytes2Hex(sha3HashWithPrefix)

	if sha3HashWithPrefixExpected != gotSha3HashWithPrefix {
		t.Errorf("Expected %s, got %s", sha3HashWithPrefixExpected, gotSha3HashWithPrefix)
	}

	msg := sha3HashWithPrefix

	priv := "b3bb9e142cac6b381060bbee24789e179beebe7fd02e93afdc45241084568240"
	key, err := crypto.HexToECDSA(priv)
	sig, err := crypto.Sign(msg, key)

	gotSig := common.Bytes2Hex(sig)
	expectedSig := "1f4ab7e26711f235331edc67bd697fd0c7628dd5ffcab333870640dee329914b2bce958fb3ee54817b1d5102e364a9164f46f732f4a02a9d5cd9569b085f211200"

	if expectedSig != gotSig {
		t.Fatalf("Expected %s, got %s", expectedSig, gotSig)
	}

	if err != nil {
		t.Fatalf("Sign error: %s", err)
	}

	R := common.Bytes2Hex(sig)[0:64]
	S := common.Bytes2Hex(sig)[64:128]
	vStr := common.Bytes2Hex(sig)[128:130]
	V, _ := strconv.Atoi(vStr)
	V = V + 27

	expected_R := "1f4ab7e26711f235331edc67bd697fd0c7628dd5ffcab333870640dee329914b"
	expected_S := "2bce958fb3ee54817b1d5102e364a9164f46f732f4a02a9d5cd9569b085f2112"
	expected_V := 27

	if expected_R != R {
		t.Errorf("Expected %s, got %s", expected_R, R)
	}

	if expected_S != S {
		t.Errorf("Expected %s, got %s", expected_S, S)
	}

	if expected_V != V {
		t.Errorf("Expected %s, got %s", expected_V, V)
	}

	recoveredPub, err := crypto.Ecrecover(msg, sig)
	pubKey := crypto.ToECDSAPub(recoveredPub)
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	addr := common.HexToAddress(orderPost.User)

	if err != nil {
		t.Fatalf("ECRecover error: %s", err)
	}

	if addr != recoveredAddr {
		t.Fatalf("Address mismatch: want: %x have: %x", addr, recoveredAddr)
	}
}

func TestGetTickerApi(t *testing.T) {
	t.Skip()
	err := GetTickerApi()

	if err != nil {

	}
}

func TestParseStringExpNotation(t *testing.T) {
	t.Skip("Skipping ParseExpNotation")

	{
		str := "2.8e+21"

		got := ParseStringExpNotation(str)
		expected := "2800000000000000000000"

		if expected != got {
			t.Errorf("Expected %s, got %s", expected, got)
		}
	}
	{
		str := "2.61-06"

		got := ParseStringExpNotation(str)
		expected := "0.00000261"

		if expected != got {
			t.Errorf("Expected %s, got %s", expected, got)
		}
	}
}
