package recording

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// WriteWAV writes mono float32 samples (range roughly -1..1) to a 16-bit PCM
// WAV file at path. It is the on-disk format for an exported take's audio.
func WriteWAV(path string, sampleRate int, samples []float32) error {
	if sampleRate <= 0 {
		return fmt.Errorf("recording: invalid sample rate %d", sampleRate)
	}
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	const (
		numChannels   = 1
		bitsPerSample = 16
	)
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataLen := len(samples) * bitsPerSample / 8

	w := func(v any) error { return binary.Write(f, binary.LittleEndian, v) }

	// RIFF header
	if _, err := f.WriteString("RIFF"); err != nil {
		return err
	}
	if err := w(uint32(36 + dataLen)); err != nil { // file size - 8
		return err
	}
	if _, err := f.WriteString("WAVE"); err != nil {
		return err
	}
	// fmt chunk
	if _, err := f.WriteString("fmt "); err != nil {
		return err
	}
	for _, v := range []any{
		uint32(16),            // PCM fmt chunk size
		uint16(1),             // audio format = PCM
		uint16(numChannels),   //
		uint32(sampleRate),    //
		uint32(byteRate),      //
		uint16(blockAlign),    //
		uint16(bitsPerSample), //
	} {
		if err := w(v); err != nil {
			return err
		}
	}
	// data chunk
	if _, err := f.WriteString("data"); err != nil {
		return err
	}
	if err := w(uint32(dataLen)); err != nil {
		return err
	}
	for _, s := range samples {
		if s > 1 {
			s = 1
		} else if s < -1 {
			s = -1
		}
		if err := w(int16(s * 32767)); err != nil {
			return err
		}
	}
	return nil
}
