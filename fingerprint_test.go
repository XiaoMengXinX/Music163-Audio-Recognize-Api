package ncmrecognize

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
)

// testdata/sine.b64 was produced by the original Node.js implementation.
func TestFingerprintMatchesNode(t *testing.T) {
	want, err := os.ReadFile("testdata/sine.b64")
	if err != nil {
		t.Fatal(err)
	}
	pcm := make([]float32, 48000*6)
	for i := range pcm {
		pcm[i] = float32(math.Sin(float64(i)*0.01) * 0.5)
	}
	got, err := Fingerprint(context.Background(), pcm, 48000, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.TrimSpace(string(want)) {
		t.Fatalf("mismatch\n got %.80s\nwant %.80s", got, want)
	}
}
