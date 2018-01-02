package helpers

import (
	"encoding/hex"
	"github.com/shopspring/decimal"
	"math/big"
	"testing"
)

func TestGetTokenDecimals(t *testing.T) {
	// BAT token
	tokenAddress := "0x0d8775f648430679a709e98d2b0cb6250d2887ef"

	decimals, err := GetTokenDecimals(tokenAddress)

	if err != nil {
		t.Errorf("Got error", err)
	}

	expected := big.NewInt(18)

	if decimals.Cmp(expected) != 0 {
		t.Errorf("Expected %s, got %s", expected, decimals)
	}
}

func TestGetLastestBlockNumber(t *testing.T) {
	blockNumber, err := GetLatestBlockNumber()

	if err != nil {
		t.Errorf("Got error", err)
	}

	if blockNumber.Cmp(big.NewInt(4000000)) != 1 {
		t.Errorf("Expected latest block number to be larger than 4,000,000, instead got %s", blockNumber)
	}
}

func TestGetSigRSV(t *testing.T) {
	sig, err := hex.DecodeString("1f4ab7e26711f235331edc67bd697fd0c7628dd5ffcab333870640dee329914b2bce958fb3ee54817b1d5102e364a9164f46f732f4a02a9d5cd9569b085f211200")

	if err != nil {
		t.Errorf("Got error %s", err)
	}

	expected_R := "0x1f4ab7e26711f235331edc67bd697fd0c7628dd5ffcab333870640dee329914b"
	expected_S := "0x2bce958fb3ee54817b1d5102e364a9164f46f732f4a02a9d5cd9569b085f2112"
	expected_V := 27

	rsv := GetSigRSV(sig)

	if expected_R != rsv.R {
		t.Errorf("Expected %s, got %s", expected_R, rsv.R)
	}

	if expected_S != rsv.S {
		t.Errorf("Expected %s, got %s", expected_S, rsv.S)
	}

	if expected_V != rsv.V {
		t.Errorf("Expected %s, got %s", expected_V, rsv.V)
	}
}

func TestDecimalsToWei(t *testing.T) {
	amount := decimal.NewFromFloat(0.02)
	got := DecimalsToWei(&amount, big.NewInt(18))
	expected := new(big.Int)
	expected.SetString("20000000000000000", 10)

	if got.Cmp(expected) != 0 {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestEthToWei(t *testing.T) {
	amountInEth := decimal.NewFromFloat(0.02)
	got := EthToWei(&amountInEth)
	expected := new(big.Int)
	expected.SetString("20000000000000000", 10)

	if got.Cmp(expected) != 0 {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}
