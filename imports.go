package ncmrecognize

import (
	"context"

	"github.com/tetratelabs/wazero/api"
)

// defineImports provides the Emscripten/embind imports of module "a".
// Embind type registration is irrelevant to us (we call raw invokers), so the
// register_* imports are no-ops; only a handful have real behaviour.
func (e *Engine) defineImports(ctx context.Context) error {
	i32 := api.ValueTypeI32
	type sig struct{ params, results int }
	sigs := map[string]sig{
		"a": {3, 0}, "b": {5, 0}, "c": {8, 0}, "d": {4, 0}, "e": {3, 0}, "f": {3, 0},
		"g": {1, 1}, "h": {0, 0}, "i": {4, 1}, "j": {3, 0}, "k": {2, 0}, "l": {6, 0},
		"m": {1, 0}, "n": {1, 0}, "o": {2, 1}, "p": {7, 0}, "q": {1, 0}, "r": {3, 1},
		"s": {1, 1}, "t": {5, 1}, "u": {2, 1}, "v": {2, 1}, "w": {6, 0}, "x": {2, 0},
		"y": {5, 0}, "z": {2, 0}, "A": {13, 0},
	}
	b := e.rt.NewHostModuleBuilder("a")
	for name, s := range sigs {
		params := make([]api.ValueType, s.params)
		for k := range params {
			params[k] = i32
		}
		results := make([]api.ValueType, s.results)
		for k := range results {
			results[k] = i32
		}
		var fn api.GoModuleFunc
		switch name {
		case "o": // _emval_take_value(type, ptr): only used for int8 reads in vector::get
			fn = func(_ context.Context, m api.Module, st []uint64) {
				v, _ := m.Memory().ReadByte(uint32(st[1]))
				e.lastVal = int8(v)
				st[0] = 5
			}
		case "r": // emscripten_memcpy_big(dst, src, n)
			fn = func(_ context.Context, m api.Module, st []uint64) {
				mem := m.Memory()
				if src, ok := mem.Read(uint32(st[1]), uint32(st[2])); ok {
					mem.Write(uint32(st[0]), append([]byte(nil), src...))
				}
				st[0] = 0
			}
		case "c": // register_class_function(cls, name, argc, argTypes, sig, invoker, memberPtr, pure)
			fn = func(_ context.Context, m api.Module, st []uint64) {
				e.methods[cstring(m, uint32(st[1]))] = uint32(st[6])
			}
		case "l": // register_function(name, argc, argTypes, sig, invoker, fn)
			fn = func(_ context.Context, m api.Module, st []uint64) {
				e.funcs[cstring(m, uint32(st[0]))] = uint32(st[5])
			}
		case "h", "d", "f", "s": // abort / assert / throw / OOM
			fn = func(context.Context, api.Module, []uint64) { panic("afp.wasm aborted") }
		default:
			fn = func(_ context.Context, _ api.Module, st []uint64) {
				for k := range st {
					st[k] = 0
				}
			}
		}
		b.NewFunctionBuilder().WithGoModuleFunction(fn, params, results).Export(name)
	}
	_, err := b.Instantiate(ctx)
	return err
}

func cstring(m api.Module, p uint32) string {
	mem := m.Memory()
	var out []byte
	for {
		b, ok := mem.ReadByte(p)
		if !ok || b == 0 {
			return string(out)
		}
		out = append(out, b)
		p++
	}
}
