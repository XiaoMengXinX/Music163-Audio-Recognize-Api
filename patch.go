package ncmrecognize

import (
	"errors"
	"math"
	"sort"
)

func mathFloat32bits(f float32) uint32 { return math.Float32bits(f) }

func readLEB(b []byte, p int) (uint32, int) {
	var r uint32
	var s uint
	for {
		x := b[p]
		p++
		r |= uint32(x&0x7f) << s
		s += 7
		if x&0x80 == 0 {
			return r, p
		}
	}
}

func appendLEB(b []byte, v uint32) []byte {
	for {
		x := byte(v & 0x7f)
		v >>= 7
		if v != 0 {
			b = append(b, x|0x80)
		} else {
			return append(b, x)
		}
	}
}

// patchExports appends function exports to the wasm export section.
func patchExports(wasm []byte, extra map[string]uint32) ([]byte, error) {
	names := make([]string, 0, len(extra))
	for k := range extra {
		names = append(names, k)
	}
	sort.Strings(names)

	p := 8
	for p < len(wasm) {
		id := wasm[p]
		size, body := readLEB(wasm, p+1)
		end := body + int(size)
		if id != 7 {
			p = end
			continue
		}
		count, q := readLEB(wasm, body)
		content := append([]byte{}, wasm[q:end]...)
		for _, n := range names {
			content = appendLEB(content, uint32(len(n)))
			content = append(content, n...)
			content = append(content, 0x00) // func
			content = appendLEB(content, extra[n])
		}
		sec := appendLEB(nil, count+uint32(len(names)))
		sec = append(sec, content...)
		out := append([]byte{}, wasm[:p]...)
		out = append(out, 7)
		out = appendLEB(out, uint32(len(sec)))
		out = append(out, sec...)
		return append(out, wasm[end:]...), nil
	}
	return nil, errors.New("export section not found")
}
