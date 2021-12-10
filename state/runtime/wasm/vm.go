package wasm

import (
	"encoding/json"
	"io/ioutil"

	"github.com/0xPolygon/polygon-sdk/chain"
	"github.com/0xPolygon/polygon-sdk/state/runtime"
	"github.com/anconprotocol/sdk"
	"github.com/ethereum/go-ethereum/common/hexutil"
	wasm_validation "github.com/perlin-network/life/wasm-validation"
	"github.com/spf13/cast"

	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

var _ runtime.Runtime = &WASM{}

type contract interface {
	gas(input []byte, config *chain.ForksInTime) uint64
	run(input []byte) ([]byte, error)
}

// WASM is the ethereum virtual machine
type WASM struct {
	engine *wasmer.Engine
	store  sdk.Storage
}

// NewEVM creates a new WASM
func NewVM(s sdk.Storage) *WASM {
	engine := wasmer.NewEngine()
	return &WASM{store: s, engine: engine}
}

// CanRun implements the runtime interface
func (e *WASM) CanRun(c *runtime.Contract, _ runtime.Host, _ *chain.ForksInTime) bool {
	if err := wasm_validation.ValidateWasm(c.Code); err != nil {
		return false
	}
	return true
}

// Name implements the runtime interface
func (e *WASM) Name() string {
	return "wasm"
}

// Run implements the runtime interface
func (e *WASM) Run(c *runtime.Contract, host runtime.Host, config *chain.ForksInTime) *runtime.ExecutionResult {

	wasmBytes, _ := ioutil.ReadFile("/home/rogelio/Code/polygon-sdk/simple.wasm")
	arr := make([]int64, 2)
	arr[0] = 5
	arr[1] = 7
	input, _ := json.Marshal(arr)
	v := hexutil.Encode(input)

	var args []interface{}
	hexbytes := hexutil.MustDecode(v)

	err := json.Unmarshal(hexbytes, &args)
	if err != nil {
		panic(err)
	}

	targs := make([]interface{}, len(args))
	for i := 0; i < len(args); i++ {
		targs[i] = cast.ToInt32(args[i])
	}

	store := wasmer.NewStore(e.engine)

	// Compiles the module
	module, _ := wasmer.NewModule(store, wasmBytes)

	// Instantiates the module
	importObject := wasmer.NewImportObject()
	instance, _ := wasmer.NewInstance(module, importObject)

	main, _ := instance.Exports.GetFunction("sum")

	// Calls that exported function with Go standard values. The WebAssembly
	// types are inferred and values are casted automatically.
	result, err := main((targs)...)

	hexvalue, _ := toHex(result)

	// gasCost := vm.GasPolicy.GetCost()

	// // In the case of not enough gas for precompiled execution we return ErrOutOfGas
	// if c.Gas < gasCost {
	// 	return &runtime.ExecutionResult{
	// 		GasLeft: 0,
	// 		Err:     runtime.ErrOutOfGas,
	// 	}
	// }

	// c.Gas = c.Gas - gasCost

	// result := &runtime.ExecutionResult{
	// 	ReturnValue: returnValue,
	// 	GasLeft:     c.Gas,
	// 	Err:         err,
	// }

	return &runtime.ExecutionResult{
		ReturnValue: hexvalue,
		// GasLeft:     gasLeft,
		Err: err,
	}
}

func toHex(result interface{}) ([]byte, error) {
	var hexresult hexutil.Bytes

	hexresult, err := json.Marshal(result)

	if err != nil {
		return nil, err
	}

	hexvalue, err := hexresult.MarshalText()

	if err != nil {
		return nil, err
	}
	return hexvalue, nil
}
