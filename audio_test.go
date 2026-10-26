package ncmrecognize

import (
	"math"
	"os"
	"testing"
)

// Decodes generated 440 Hz stereo files in every supported format and checks
// the dominant frequency (via zero crossings) and that fingerprinting works.
func TestDecodeFormats(t *testing.T) {
	for _, f := range []struct {
		file string
		rate int
	}{
		{"sine440.mp3", 44100},
		{"sine440.wav", 44100},
		{"sine440.flac", 44100},
		{"sine440_vorbis.ogg", 44100},
		{"sine440_opus.ogg", 48000},
	} {
		t.Run(f.file, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + f.file)
			if err != nil {
				t.Fatal(err)
			}
			a, err := Decode(data)
			if err != nil {
				t.Fatal(err)
			}
			if a.SampleRate != f.rate || len(a.PCM) < f.rate*7 {
				t.Fatalf("rate=%d len=%d", a.SampleRate, len(a.PCM))
			}
			z := 0
			seg := a.PCM[f.rate : 2*f.rate]
			for i := 1; i < len(seg); i++ {
				if (seg[i-1] < 0) != (seg[i] < 0) {
					z++
				}
			}
			if math.Abs(float64(z)-880) > 10 {
				t.Fatalf("zero crossings %d, want ~880", z)
			}
			if _, err := Fingerprint(t.Context(), a.PCM, a.SampleRate, 1); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDecodeUnsupported(t *testing.T) {
	if _, err := Decode([]byte("\x00\x00\x00\x18ftypM4A not supported")); err == nil {
		t.Fatal("expected error")
	}
}
