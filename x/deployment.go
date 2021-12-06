package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/0xPolygon/polygon-sdk/helper/tests"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/jsonrpc"
)

func main() {

	// OnchainMetadata
	contract := "608060405234801561001057600080fd5b506004361061002b5760003560e01c8063a8f8088214610030575b600080fd5b61004361003e36600461015e565b610045565b005b836040516100539190610283565b60405180910390207ff65e7a14beca79936be9e3bb8ae5c826723a0e8a8d9ffcd6c7f987a0734f9c9387878686866040516100929594939291906102cb565b60405180910390a2505050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600082601f8301126100e257600080fd5b813567ffffffffffffffff808211156100fd576100fd6100a2565b604051601f8301601f19908116603f01168101908282118183101715610125576101256100a2565b8160405283815286602085880101111561013e57600080fd5b836020870160208301376000602085830101528094505050505092915050565b60008060008060008060c0878903121561017757600080fd5b863567ffffffffffffffff8082111561018f57600080fd5b61019b8a838b016100d1565b975060208901359150808211156101b157600080fd5b6101bd8a838b016100d1565b965060408901359150808211156101d357600080fd5b6101df8a838b016100d1565b955060608901359150808211156101f557600080fd5b6102018a838b016100d1565b9450608089013591508082111561021757600080fd5b6102238a838b016100d1565b935060a089013591508082111561023957600080fd5b5061024689828a016100d1565b9150509295509295509295565b60005b8381101561026e578181015183820152602001610256565b8381111561027d576000848401525b50505050565b60008251610295818460208701610253565b9190910192915050565b600081518084526102b7816020860160208601610253565b601f01601f19169290920160200192915050565b60a0815260006102de60a083018861029f565b82810360208401526102f0818861029f565b90508281036040840152610304818761029f565b90508281036060840152610318818661029f565b9050828103608084015261032c818561029f565b9897505050505050505056fea264697066735822122010fdcbd553b376acca00bf513d063d756de33941b4fecfb1e1363db75439dca764736f6c634300080a0033"

	addr, err := DeployContract(context.Background(), contract)
	if err != nil {
		panic(err)
	}

	fmt.Println(addr)
}

// DeployContract deploys a contract with account 0 and returns the address
func DeployContract(ctx context.Context, binary string) (web3.Address, error) {
	buf, err := hex.DecodeString(binary)
	if err != nil {
		return web3.Address{}, err
	}

	// deploy := &web3.Transaction{
	// 	Gas:      framework.DefaultGasLimit,
	// 	GasPrice: framework.DefaultGasPrice,
	// 	Value:    big.NewInt(0),
	// 	Input:    buf,
	// }

	receipt, err := SendTxn(ctx, &web3.Transaction{
		From:  web3.HexToAddress("0x32A21c1bB6E7C20F547e930b53dAC57f42cd25F6"),
		Input: buf,
	})
	if err != nil {
		return web3.Address{}, err
	}
	return receipt.ContractAddress, nil
}

const (
	DefaultGasPrice = 1879048192 // 0x70000000
	DefaultGasLimit = 5242880    // 0x500000
)

var emptyAddr web3.Address

func SendTxn(ctx context.Context, txn *web3.Transaction) (*web3.Receipt, error) {
	client, err := jsonrpc.NewClient("http://localhost:8545")
	if err != nil {
		panic(err)
	}

	// if txn.From == emptyAddr {
	// 	txn.From = web3.Address(t.Config.PremineAccts[0].Addr)
	// }
	if txn.GasPrice == 0 {
		txn.GasPrice = DefaultGasPrice
	}
	if txn.Gas == 0 {
		txn.Gas = DefaultGasLimit
	}
	hash, err := client.Eth().SendTransaction(txn)
	if err != nil {
		return nil, err
	}
	return WaitForReceipt(ctx, hash)
}

// // SendRawTx signs the transaction with the provided private key, executes it, and returns the receipt
// func SendRawTx(ctx context.Context, tx *PreparedTransaction, signerKey *ecdsa.PrivateKey) (*web3.Receipt, error) {
// 	signer := crypto.NewEIP155Signer(100)
// 	client := t.JSONRPC()

// 	nextNonce, err := client.Eth().GetNonce(web3.Address(tx.From), web3.Latest)
// 	if err != nil {
// 		return nil, err
// 	}

// 	signedTx, err := signer.SignTx(&types.Transaction{
// 		From:     tx.From,
// 		GasPrice: tx.GasPrice,
// 		Gas:      tx.Gas,
// 		To:       tx.To,
// 		Value:    tx.Value,
// 		Input:    tx.Input,
// 		Nonce:    nextNonce,
// 	}, signerKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	txHash, err := client.Eth().SendRawTransaction(signedTx.MarshalRLP())
// 	if err != nil {
// 		return nil, err
// 	}

// 	return t.WaitForReceipt(ctx, txHash)
// }

func WaitForReceipt(ctx context.Context, hash web3.Hash) (*web3.Receipt, error) {
	client, err := jsonrpc.NewClient("http://localhost:8545")
	if err != nil {
		panic(err)
	}

	type result struct {
		receipt *web3.Receipt
		err     error
	}

	res, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
		receipt, err := client.Eth().GetTransactionReceipt(hash)
		if err != nil && err.Error() != "not found" {
			return result{receipt, err}, false
		}
		if receipt != nil {
			return result{receipt, nil}, false
		}
		return nil, true
	})
	if err != nil {
		return nil, err
	}
	data := res.(result)
	return data.receipt, data.err
}

// func WaitForReady(ctx context.Context) error {
// 	_, err := tests.RetryUntilTimeout(ctx, func() (interface{}, bool) {
// 		num, err := t.GetLatestBlockHeight()
// 		if err != nil {
// 			return nil, true
// 		}
// 		if num == 0 {
// 			return nil, true
// 		}
// 		return num, false
// 	})
// 	return err
// }

// func TxnTo(ctx context.Context, address web3.Address, method string) *web3.Receipt {
// 	sig := MethodSig(method)
// 	receipt, err := t.SendTxn(ctx, &web3.Transaction{
// 		To:    &address,
// 		Input: sig,
// 	})
// 	if err != nil {
// 		t.t.Fatal(err)
// 	}
// 	return receipt
// }
