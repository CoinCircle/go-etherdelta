package helpers

import (
	"context"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	token "github.com/miguelmota/go-etherdelta/helpers/token"
	"github.com/shopspring/decimal"
)

func DecimalsToWei(amount *decimal.Decimal, decimals *big.Int) *big.Int {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromBigInt(decimals, 0))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}

func EthToWei(amount *decimal.Decimal) *big.Int {
	wei := DecimalsToWei(amount, big.NewInt(18))
	return wei
}

func GetAccountBalance(address string) (*big.Int, error) {
	accountAddress := common.HexToAddress(address)
	rawBalance, err := Client.BalanceAt(context.Background(), accountAddress, nil)

	if err != nil {
		return nil, err
	}

	return rawBalance, nil
}

func GetTokenBalance(_tokenAddress string, _accountAddress string) (*big.Int, error) {
	var bal *big.Int
	tokenAddress := common.HexToAddress(_tokenAddress)
	accountAddress := common.HexToAddress(_accountAddress)

	instance, err := token.NewTokenCaller(tokenAddress, Client)

	if err != nil {
		return bal, err
	}

	bal, err = instance.BalanceOf(&bind.CallOpts{Pending: false}, accountAddress)

	if err != nil {
		return bal, err
	}

	return bal, nil
}

func GetTokenAllowance(_tokenAddress string, _accountAddress string, _spenderAddress string) (*big.Int, error) {
	var allowed *big.Int
	tokenAddress := common.HexToAddress(_tokenAddress)
	accountAddress := common.HexToAddress(_accountAddress)
	spenderAddress := common.HexToAddress(_spenderAddress)

	instance, err := token.NewTokenCaller(tokenAddress, Client)

	if err != nil {
		return allowed, err
	}

	allowed, err = instance.Allowance(&bind.CallOpts{Pending: false}, accountAddress, spenderAddress)

	if err != nil {
		return allowed, err
	}

	return allowed, nil
}

func GetTokenDecimals(tokenAddress string) (*big.Int, error) {
	var decimals *big.Int

	instance, err := token.NewTokenCaller(common.HexToAddress(tokenAddress), Client)

	if err != nil {
		return decimals, err
	}

	decimalsInt8, err := instance.Decimals(&bind.CallOpts{})

	decimals = big.NewInt(int64(decimalsInt8))

	if err != nil {
		return decimals, err
	}

	return decimals, nil
}

func TransferTokens(auth bind.TransactOpts, _tokenAddress string, _toAddress string, amount *big.Int) (*types.Transaction, error) {
	tokenAddress := common.HexToAddress(_tokenAddress)
	toAddress := common.HexToAddress(_toAddress)
	instance, err := token.NewToken(tokenAddress, Client)

	if err != nil {
		return &types.Transaction{}, err
	}

	tx, err := instance.Transfer(&auth, toAddress, amount)

	if err != nil {
		return tx, err
	}

	return tx, nil
}

func GetLatestBlockNumber() (*big.Int, error) {
	block, err := Client.HeaderByNumber(context.Background(), nil)

	return block.Number, err
}

func GetSigRSV(sig []byte) *RSV {
	R := "0x" + common.Bytes2Hex(sig)[0:64]
	S := "0x" + common.Bytes2Hex(sig)[64:128]
	vStr := common.Bytes2Hex(sig)[128:130]
	V, _ := strconv.Atoi(vStr)
	V = V + 27

	rsv := &RSV{R, S, V}

	return rsv
}

func GetSigRSVBytes(sig []byte) ([32]byte, [32]byte, uint8) {
	sigstr := common.Bytes2Hex(sig)
	_R := sigstr[0:64]
	_S := sigstr[64:128]

	R := [32]byte{}
	S := [32]byte{}

	copy(R[:], common.FromHex(_R))
	copy(S[:], common.FromHex(_S))

	vStr := sigstr[128:130]
	_V, _ := strconv.Atoi(vStr)
	V := uint8(_V + 27)

	return R, S, V
}
