package jsonrpc

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-sdk/helper/hex"
	"github.com/0xPolygon/polygon-sdk/state"
	"github.com/0xPolygon/polygon-sdk/types"
)

// Ancon is the ancon jsonrpc endpoint
type Ancon struct {
	d *Dispatcher
}

// GetBlockByHash returns information about a block by hash
func (a *Ancon) GetBlockByHash(hash types.Hash, fullTx bool) (interface{}, error) {
	block, ok := a.d.store.GetBlockByHash(hash, true)
	if !ok {
		return nil, nil
	}
	return toBlock(block, fullTx), nil
}

// BlockNumber returns current block number
func (a *Ancon) BlockNumber() (interface{}, error) {
	h := a.d.store.Header()
	if h == nil {
		return nil, fmt.Errorf("header has a nil value")
	}
	return argUintPtr(h.Number), nil
}

// SendRawTransaction sends a raw transaction
func (a *Ancon) SendRawTransaction(input string) (interface{}, error) {
	buf := hex.MustDecodeHex(input)

	tx := &types.Transaction{}
	if err := tx.UnmarshalRLP(buf); err != nil {
		return nil, err
	}
	tx.ComputeHash()

	if err := a.d.store.AddTx(tx); err != nil {
		return nil, err
	}
	return tx.Hash.String(), nil
}

// SendTransaction creates new message call transaction or a contract creation, if the data field contains code.
func (a *Ancon) SendTransaction(arg *txnArgs) (interface{}, error) {
	transaction, err := a.d.decodeTxn(arg)
	if err != nil {
		return nil, err
	}
	if err := a.d.store.AddTx(transaction); err != nil {
		return nil, err
	}
	return transaction.Hash.String(), nil
}

// GetTransactionByHash returns a transaction by his hash
func (a *Ancon) GetTransactionByHash(hash types.Hash) (interface{}, error) {
	blockHash, ok := a.d.store.ReadTxLookup(hash)
	if !ok {
		// txn not found
		return nil, nil
	}
	block, ok := a.d.store.GetBlockByHash(blockHash, true)
	if !ok {
		// block receipts not found
		return nil, nil
	}
	for idx, txn := range block.Transactions {
		if txn.Hash == hash {
			return toTransaction(txn, block, idx), nil
		}
	}
	// txn not found (this should not happen)
	a.d.logger.Warn(
		fmt.Sprintf("Transaction with hash [%s] not found", blockHash),
	)
	return nil, nil
}

// GetTransactionReceipt returns a transaction receipt by his hash
func (a *Ancon) GetTransactionReceipt(hash types.Hash) (interface{}, error) {
	blockHash, ok := a.d.store.ReadTxLookup(hash)
	if !ok {
		// txn not found
		return nil, nil
	}

	block, ok := a.d.store.GetBlockByHash(blockHash, true)
	if !ok {
		// block not found
		a.d.logger.Warn(
			fmt.Sprintf("Block with hash [%s] not found", blockHash.String()),
		)
		return nil, nil
	}

	receipts, err := a.d.store.GetReceiptsByHash(blockHash)
	if err != nil {
		// block receipts not found
		a.d.logger.Warn(
			fmt.Sprintf("Receipts for block with hash [%s] not found", blockHash.String()),
		)
		return nil, nil
	}
	if len(receipts) == 0 {
		// Receipts not written yet on the db
		a.d.logger.Warn(
			fmt.Sprintf("No receipts found for block with hash [%s]", blockHash.String()),
		)
		return nil, nil
	}
	// find the transaction in the body
	indx := -1
	for i, txn := range block.Transactions {
		if txn.Hash == hash {
			indx = i
			break
		}
	}
	if indx == -1 {
		// txn not found
		return nil, nil
	}

	txn := block.Transactions[indx]
	raw := receipts[indx]

	logs := make([]*Log, len(raw.Logs))
	for indx, elem := range raw.Logs {
		logs[indx] = &Log{
			Address:     elem.Address,
			Topics:      elem.Topics,
			Data:        argBytes(elem.Data),
			BlockHash:   block.Hash(),
			BlockNumber: argUint64(block.Number()),
			TxHash:      txn.Hash,
			TxIndex:     argUint64(indx),
			LogIndex:    argUint64(indx),
			Removed:     false,
		}
	}
	res := &receipt{
		Root:              raw.Root,
		CumulativeGasUsed: argUint64(raw.CumulativeGasUsed),
		LogsBloom:         raw.LogsBloom,
		Status:            argUint64(*raw.Status),
		TxHash:            txn.Hash,
		TxIndex:           argUint64(indx),
		BlockHash:         block.Hash(),
		BlockNumber:       argUint64(block.Number()),
		GasUsed:           argUint64(raw.GasUsed),
		ContractAddress:   raw.ContractAddress,
		FromAddr:          txn.From,
		ToAddr:            txn.To,
		Logs:              logs,
	}
	return res, nil
}

// GasPrice returns the average gas price based on the last x blocks
func (a *Ancon) GasPrice() (interface{}, error) {
	// Grab the average gas price and convert it to a hex value
	avgGasPrice := hex.EncodeBig(a.d.store.GetAvgGasPrice())

	return avgGasPrice, nil
}

// Call executes a smart contract call using the transaction object data
func (a *Ancon) Call(
	arg *txnArgs,
	number *BlockNumber,
) (interface{}, error) {
	if number == nil {
		number, _ = createBlockNumberPointer("latest")
	}
	transaction, err := a.d.decodeTxn(arg)
	if err != nil {
		return nil, err
	}
	// Fetch the requested header
	header, err := a.d.getBlockHeaderImpl(*number)
	if err != nil {
		return nil, err
	}

	// If the caller didn't supply the gas limit in the message, then we set it to maximum possible => block gas limit
	if transaction.Gas == 0 {
		transaction.Gas = header.GasLimit
	}

	// The return value of the execution is saved in the transition (returnValue field)
	result, err := a.d.store.ApplyTxn(header, transaction)
	if err != nil {
		return nil, err
	}

	if result.Failed() {
		return nil, fmt.Errorf("unable to execute call")
	}
	return argBytesPtr(result.ReturnValue), nil
}

// EstimateGas estimates the gas needed to execute a transaction
func (a *Ancon) EstimateGas(
	arg *txnArgs,
	rawNum *BlockNumber,
) (interface{}, error) {
	transaction, err := a.d.decodeTxn(arg)
	if err != nil {
		return nil, err
	}

	number := LatestBlockNumber
	if rawNum != nil {
		number = *rawNum
	}

	// Fetch the requested header
	header, err := a.d.getBlockHeaderImpl(number)
	if err != nil {
		return nil, err
	}

	forksInTime := a.d.store.GetForksInTime(uint64(number))

	var standardGas uint64
	if transaction.IsContractCreation() && forksInTime.Homestead {
		standardGas = state.TxGasContractCreation
	} else {
		standardGas = state.TxGas
	}

	var (
		lowEnd  = standardGas
		highEnd uint64
		gasCap  uint64
	)

	// If the gas limit was passed in, use it as a ceiling
	if transaction.Gas != 0 && transaction.Gas >= standardGas {
		highEnd = transaction.Gas
	} else {
		// If not, use the referenced block number
		highEnd = header.GasLimit
	}

	gasPriceInt := new(big.Int).Set(transaction.GasPrice)
	valueInt := new(big.Int).Set(transaction.Value)

	// If the sender address is present, recalculate the ceiling to his balance
	if transaction.From != types.ZeroAddress && transaction.GasPrice != nil && gasPriceInt.BitLen() != 0 {
		// Get the account balance

		// If the account is not initialized yet in state,
		// assume it's an empty account
		accountBalance := big.NewInt(0)
		acc, err := a.d.store.GetAccount(header.StateRoot, transaction.From)
		if err != nil && !errors.As(err, &ErrStateNotFound) {
			// An unrelated error occurred, return it
			return nil, err
		} else if err == nil {
			// No error when fetching the account,
			// read the balance from state
			accountBalance = acc.Balance
		}

		available := new(big.Int).Set(accountBalance)

		if transaction.Value != nil {
			if valueInt.Cmp(available) >= 0 {
				return nil, fmt.Errorf("insufficient funds for execution")
			}

			available.Sub(available, valueInt)
		}

		allowance := new(big.Int).Div(available, gasPriceInt)

		// If the allowance is larger than maximum uint64, skip checking
		if allowance.IsUint64() && highEnd > allowance.Uint64() {
			highEnd = allowance.Uint64()
		}
	}

	if highEnd > types.GasCap.Uint64() {
		// The high end is greater than the environment gas cap
		highEnd = types.GasCap.Uint64()
	}

	gasCap = highEnd

	// Run the transaction with the estimated gas
	testTransaction := func(gas uint64) (bool, error) {
		// Create a dummy transaction with the new gas
		txn := transaction.Copy()
		txn.Gas = gas

		result, err := a.d.store.ApplyTxn(header, txn)

		if err != nil {
			return true, err
		}

		return result.Failed(), nil
	}

	// Start the binary search for the lowest possible gas price
	for lowEnd <= highEnd {
		mid := (lowEnd + highEnd) / 2

		failed, err := testTransaction(mid)
		if err != nil {
			return 0, err
		}

		if failed {
			// If the transaction failed => increase the gas
			lowEnd = mid + 1
		} else {
			// If the transaction didn't fail => lower the gas
			highEnd = mid - 1
		}
	}

	// we stopped the binary search at the last gas limit
	// at which the txn could not be executed
	highEnd++

	// Check the edge case if even the highest cap is not enough to complete the transaction
	if highEnd == gasCap {
		failed, err := testTransaction(gasCap)

		if err != nil {
			return 0, err
		}

		if failed {
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", gasCap)
		}
	}

	return hex.EncodeUint64(highEnd), nil
}
