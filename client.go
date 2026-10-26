package ncmrecognize

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const matchURL = "https://interface.music.163.com/api/music/audio/match"

// Client calls the NetEase audio match API.
type Client struct {
	HTTP   *http.Client // defaults to http.DefaultClient
	Engine *Engine      // defaults to a shared engine
}

// Response is the API response. Raw always holds the unmodified JSON.
type Response struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
	Raw  json.RawMessage `json:"-"`
}

// Recognize fingerprints the audio (wav/mp3 bytes) and queries the API.
func (c *Client) Recognize(ctx context.Context, audio []byte) (*Response, error) {
	a, err := Decode(audio)
	if err != nil {
		return nil, err
	}
	return c.RecognizePCM(ctx, a.PCM, a.SampleRate, 0)
}

// RecognizeFile is Recognize for a file path.
func (c *Client) RecognizeFile(ctx context.Context, path string) (*Response, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return c.Recognize(ctx, b)
}

// RecognizePCM queries the API with mono PCM samples (first channel) starting at offsetSec.
func (c *Client) RecognizePCM(ctx context.Context, pcm []float32, sampleRate int, offsetSec float64) (*Response, error) {
	var fp string
	var err error
	if c.Engine != nil {
		fp, err = c.Engine.Fingerprint(ctx, pcm, sampleRate, offsetSec)
	} else {
		fp, err = Fingerprint(ctx, pcm, sampleRate, offsetSec)
	}
	if err != nil {
		return nil, err
	}
	return c.Match(ctx, fp)
}

// Match sends an already computed fingerprint to the API.
func (c *Client) Match(ctx context.Context, fingerprint string) (*Response, error) {
	form := url.Values{
		"sessionId":     {newUUID()},
		"algorithmCode": {"shazam_v2"},
		"duration":      {fmt.Sprint(FingerprintSeconds)},
		"rawdata":       {fingerprint},
		"times":         {"2"},
		"decrypt":       {"1"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, matchURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	for k, v := range map[string]string{
		"accept":             "*/*",
		"accept-language":    "zh-CN,zh-TW;q=0.9,zh;q=0.8",
		"cache-control":      "no-cache",
		"content-type":       "application/x-www-form-urlencoded; charset=UTF-8",
		"pragma":             "no-cache",
		"sec-ch-ua-platform": `"Windows"`,
		"sec-fetch-dest":     "empty",
		"sec-fetch-mode":     "cors",
		"sec-fetch-site":     "none",
	} {
		req.Header.Set(k, v)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &Response{Raw: body}
	if err := json.Unmarshal(body, r); err != nil {
		return nil, fmt.Errorf("bad response (HTTP %d): %w", resp.StatusCode, err)
	}
	return r, nil
}

func newUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
