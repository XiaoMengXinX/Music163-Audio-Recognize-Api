package ncmrecognize

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/kazzmir/opus-go/ogg"
	"github.com/kazzmir/opus-go/opus"
	"github.com/mewkiz/flac"
)

// Audio is decoded audio: the first channel as float32 in [-1, 1].
type Audio struct {
	PCM        []float32
	SampleRate int
}

// Decode decodes audio data, detecting the format by content.
//
// Supported: WAV (PCM), MP3, FLAC, Ogg Vorbis, Ogg Opus. Anything else
// (AAC/M4A, WMA, ...) must be decoded by the caller and passed as PCM to
// Fingerprint / Client.RecognizePCM.
func Decode(data []byte) (*Audio, error) {
	switch {
	case len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		return decodeWAV(data)
	case len(data) > 4 && string(data[:4]) == "fLaC":
		return decodeFLAC(data)
	case len(data) > 4 && string(data[:4]) == "OggS":
		return decodeOgg(data)
	case len(data) > 3 && (string(data[:3]) == "ID3" || (data[0] == 0xff && data[1]&0xe0 == 0xe0)):
		return decodeMP3(data)
	}
	return nil, errors.New("unsupported audio format (supported: wav, mp3, flac, ogg vorbis/opus); decode it yourself and use Fingerprint/RecognizePCM")
}

func decodeWAV(data []byte) (*Audio, error) {
	d := wav.NewDecoder(bytes.NewReader(data))
	buf, err := d.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("decode wav: %w", err)
	}
	ch := buf.Format.NumChannels
	if ch < 1 {
		return nil, errors.New("decode wav: no channels")
	}
	scale := float32(int64(1) << (buf.SourceBitDepth - 1))
	if buf.SourceBitDepth == 8 { // unsigned 8-bit
		scale = 128
	}
	n := len(buf.Data) / ch
	pcm := make([]float32, n)
	for i := range pcm {
		v := float32(buf.Data[i*ch])
		if buf.SourceBitDepth == 8 {
			v -= 128
		}
		pcm[i] = v / scale
	}
	return &Audio{PCM: pcm, SampleRate: buf.Format.SampleRate}, nil
}

func decodeMP3(data []byte) (*Audio, error) {
	d, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode mp3: %w", err)
	}
	raw, err := io.ReadAll(d) // 16-bit LE stereo
	if err != nil {
		return nil, fmt.Errorf("decode mp3: %w", err)
	}
	pcm := make([]float32, len(raw)/4)
	for i := range pcm {
		pcm[i] = float32(int16(uint16(raw[i*4])|uint16(raw[i*4+1])<<8)) / 32768
	}
	return &Audio{PCM: pcm, SampleRate: d.SampleRate()}, nil
}

func decodeFLAC(data []byte) (*Audio, error) {
	st, err := flac.New(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode flac: %w", err)
	}
	scale := float32(int64(1) << (st.Info.BitsPerSample - 1))
	var pcm []float32
	for {
		fr, err := st.ParseNext()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode flac: %w", err)
		}
		for _, v := range fr.Subframes[0].Samples { // first channel
			pcm = append(pcm, float32(v)/scale)
		}
	}
	return &Audio{PCM: pcm, SampleRate: int(st.Info.SampleRate)}, nil
}

// decodeOgg handles Ogg Opus and Ogg Vorbis, told apart by the first packet.
func decodeOgg(data []byte) (*Audio, error) {
	if bytes.Contains(data[:min(len(data), 128)], []byte("OpusHead")) {
		return decodeOpus(data)
	}
	samples, f, err := oggvorbis.ReadAll(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode ogg vorbis: %w", err)
	}
	return &Audio{PCM: firstChannel(samples, f.Channels), SampleRate: f.SampleRate}, nil
}

func decodeOpus(data []byte) (*Audio, error) {
	r, err := ogg.NewOpusReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode opus: %w", err)
	}
	dec, err := opus.NewDecoderFromHead(r.Head)
	if err != nil {
		return nil, fmt.Errorf("decode opus: %w", err)
	}
	ch := int(r.Head.Channels)
	var all []float32
	for {
		pkt, err := r.ReadAudioPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("decode opus: %w", err)
		}
		pcm, n, err := dec.DecodePacketF32(pkt, nil)
		if err != nil {
			return nil, fmt.Errorf("decode opus: %w", err)
		}
		all = append(all, pcm[:n*ch]...)
	}
	pcm := firstChannel(all, ch)
	if skip := int(r.Head.PreSkip); skip < len(pcm) { // encoder delay
		pcm = pcm[skip:]
	}
	return &Audio{PCM: pcm, SampleRate: 48000}, nil // Opus always decodes at 48 kHz
}

// firstChannel extracts channel 0 from interleaved samples.
func firstChannel(interleaved []float32, ch int) []float32 {
	if ch <= 1 {
		return interleaved
	}
	out := make([]float32, len(interleaved)/ch)
	for i := range out {
		out[i] = interleaved[i*ch]
	}
	return out
}
