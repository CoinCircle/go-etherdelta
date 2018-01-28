abigen:
	abigen --abi resources/erc20.abi --pkg etherdelta --type ERC20 --out contracts/erc20.go
	abigen --abi resources/etherdelta.abi --pkg etherdelta --type EtherDelta --out contracts/etherdelta.go
	abigen --abi resources/ethertoken.abi --pkg etherdelta --type EtherToken --out contracts/ethertoken.go
	abigen --abi resources/erc20.abi --pkg etherdelta --type Token --out helpers/token/erc20.go
