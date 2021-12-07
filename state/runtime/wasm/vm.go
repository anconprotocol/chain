package wasm

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/0xPolygon/polygon-sdk/chain"
	"github.com/0xPolygon/polygon-sdk/state/runtime"
	"github.com/perlin-network/life/exec"
	wasm_validation "github.com/perlin-network/life/wasm-validation"
	"github.com/spf13/cast"
)

var _ runtime.Runtime = &WASM{}

// WASM is the ethereum virtual machine
type WASM struct {
}

// NewEVM creates a new WASM
func NewVM() *WASM {
	return &WASM{}
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

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(c.Code, exec.VMConfig{
		DefaultMemoryPages: 128,
		DefaultTableSize:   65536,
		// DisableFloatingPoint: *noFloatingPointFlag,
	}, nil, nil)

	if err != nil {
		panic(err)
	}

	// if *pmFlag {
	// 	compileStartTime := time.Now()
	// 	fmt.Println("[Polymerase] Compilation started.")
	// 	aotSvc := platform.FullAOTCompile(vm)
	// 	if aotSvc != nil {
	// 		compileEndTime := time.Now()
	// 		fmt.Printf("[Polymerase] Compilation finished successfully in %+v.\n", compileEndTime.Sub(compileStartTime))
	// 		vm.SetAOTService(aotSvc)
	// 	} else {
	// 		fmt.Println("[Polymerase] The current platform is not yet supported.")
	// 	}
	// }

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport("main")
	if !ok {
		fmt.Printf("Entry function %s not found; starting from 0.\n", "main")
		entryID = 0
	}

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}
	var args []int64
	in := cast.ToInt64(c.Input)
	for _, arg := range flag.Args()[1:] {
		fmt.Println(arg)
		if ia, err := strconv.Atoi(arg); err != nil {
			panic(err)
		} else {
			args = append(args, int64(ia))
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err := vm.Run(entryID, in)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	return &runtime.ExecutionResult{
		ReturnValue: []byte(cast.ToString(ret)),
		// GasLeft:     gasLeft,
		Err: err,
	}
}
