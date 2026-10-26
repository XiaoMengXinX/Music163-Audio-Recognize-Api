package ncmrecognize

import (
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

//go:embed afp.wasm
var afpWasm []byte

// Function indices inside afp.wasm (reverse engineered). They are exported
// at load time by patchExports so the host can call the embind raw invokers
// directly, without implementing the embind runtime.
const (
	funcInvokeExtract = 507 // invoker(fnTableIdx, std::string wire) -> vector<char>*
	funcInvokeSize    = 574 // invoker(memberPtr, this) -> size
	funcInvokeGet     = 569 // invoker(memberPtr, this, idx) -> emval handle
	funcDestroy       = 388 // vector<char> destructor(ptr)
)

// Engine runs the fingerprint algorithm. It is safe for concurrent use.
type Engine struct {
	mu      sync.Mutex
	rt      wazero.Runtime
	mod     api.Module
	lastVal int8
	// captured from embind registration during __wasm_call_ctors
	methods map[string]uint32 // vector<char> method name -> member pointer data
	funcs   map[string]uint32 // free function name -> table slot
}

// NewEngine compiles and instantiates the embedded afp.wasm.
func NewEngine(ctx context.Context) (*Engine, error) {
	patched, err := patchExports(afpWasm, map[string]uint32{
		"invoke_extract": funcInvokeExtract,
		"invoke_size":    funcInvokeSize,
		"invoke_get":     funcInvokeGet,
		"destroy":        funcDestroy,
	})
	if err != nil {
		return nil, err
	}
	e := &Engine{rt: wazero.NewRuntime(ctx), methods: map[string]uint32{}, funcs: map[string]uint32{}}
	if err := e.defineImports(ctx); err != nil {
		return nil, err
	}
	e.mod, err = e.rt.InstantiateWithConfig(ctx, patched, wazero.NewModuleConfig().WithName("afp"))
	if err != nil {
		e.rt.Close(ctx)
		return nil, fmt.Errorf("instantiate afp.wasm: %w", err)
	}
	if _, err := e.mod.ExportedFunction("C").Call(ctx); err != nil { // __wasm_call_ctors
		e.rt.Close(ctx)
		return nil, err
	}
	if e.funcs["ExtractQueryFP"] == 0 || e.methods["size"] == 0 || e.methods["get"] == 0 {
		e.rt.Close(ctx)
		return nil, errors.New("afp.wasm: embind registration not found")
	}
	return e, nil
}

// Close releases the wasm runtime.
func (e *Engine) Close(ctx context.Context) error { return e.rt.Close(ctx) }

// ExtractQueryFP computes the raw fingerprint for 48000 samples of 8 kHz mono PCM.
func (e *Engine) ExtractQueryFP(ctx context.Context, pcm []float32) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	mem := e.mod.Memory()
	call := func(name string, args ...uint64) (uint64, error) {
		r, err := e.mod.ExportedFunction(name).Call(ctx, args...)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", name, err)
		}
		if len(r) == 0 {
			return 0, nil
		}
		return uint64(uint32(r[0])), nil
	}

	// std::string wire format: [u32 len][bytes][NUL]
	n := uint32(len(pcm) * 4)
	ptr, err := call("E", uint64(4+n+1))
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 4+n+1)
	binary.LittleEndian.PutUint32(buf, n)
	for i, f := range pcm {
		binary.LittleEndian.PutUint32(buf[4+i*4:], mathFloat32bits(f))
	}
	if !mem.Write(uint32(ptr), buf) {
		return nil, errors.New("wasm memory write out of range")
	}
	defer call("F", ptr)

	vec, err := call("invoke_extract", uint64(e.funcs["ExtractQueryFP"]), ptr)
	if err != nil {
		return nil, err
	}
	defer call("destroy", vec)

	size, err := call("invoke_size", uint64(e.methods["size"]), vec)
	if err != nil {
		return nil, err
	}
	out := make([]byte, size)
	for i := range out {
		if _, err := call("invoke_get", uint64(e.methods["get"]), vec, uint64(i)); err != nil {
			return nil, err
		}
		out[i] = byte(e.lastVal)
	}
	return out, nil
}
