package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	extism "github.com/extism/go-sdk"
)

type AddInput struct {
	Left  int32 `json:"left"`
	Right int32 `json:"right"`
}

func main() {
	ctx := context.Background()

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: "./add_fun.wasm",
			},
		},
	}

	config := extism.PluginConfig{}
	config.EnableWasi = true

	add_num_host := extism.NewHostFunctionWithStack("sum_from_host",
		func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {

			fmt.Println("stack:", stack)
			// -> stack: [110 130]

			a := int64(stack[0])
			b := int64(stack[1])

			fmt.Println("a:", a)
			fmt.Println("b:", b)
			// -> a: 97
			// -> b: 117

			sum := a + b

			buf := new(bytes.Buffer)
			err := binary.Write(buf, binary.LittleEndian, sum)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			sum_bytes := buf.Bytes()

			mem, err := p.WriteBytes(sum_bytes)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			stack[0] = mem

		},

		[]extism.ValueType{extism.ValueTypeI64, extism.ValueTypeI64},
		[]extism.ValueType{extism.ValueTypeI64},
	)

	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{add_num_host})
	if err != nil {
		fmt.Println(err)
		return
	}

	input := AddInput{
		Left:  10,
		Right: 5,
	}

	json_input, err := json.Marshal(input)
	if err != nil {
		fmt.Println(err)
		return
	}

	exit, result, err := plugin.Call("add_from_host", json_input)

	if err != nil {
		fmt.Println(err)
		os.Exit(int(exit))
	}

	var parsed int64
	res := bytes.NewReader(result)
	err = binary.Read(res, binary.LittleEndian, &parsed)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(result)
}
