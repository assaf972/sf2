package recording

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S24-T01: WriteWAV emits a well-formed 16-bit mono PCM WAV whose header matches
// the sample rate and whose data chunk holds one int16 per sample.
func TestWriteWAVHeaderAndData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "take.wav")
	samples := []float32{0, 0.5, -0.5, 1, -1}
	require.NoError(t, WriteWAV(path, 48000, samples))

	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(b), 44, "must have a 44-byte header")
	assert.Equal(t, "RIFF", string(b[0:4]))
	assert.Equal(t, "WAVE", string(b[8:12]))
	assert.Equal(t, "fmt ", string(b[12:16]))
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(b[20:22]), "PCM format")
	assert.Equal(t, uint16(1), binary.LittleEndian.Uint16(b[22:24]), "mono")
	assert.Equal(t, uint32(48000), binary.LittleEndian.Uint32(b[24:28]), "sample rate")
	assert.Equal(t, uint16(16), binary.LittleEndian.Uint16(b[34:36]), "16-bit")
	assert.Equal(t, "data", string(b[36:40]))
	dataLen := binary.LittleEndian.Uint32(b[40:44])
	assert.Equal(t, uint32(len(samples)*2), dataLen, "2 bytes per mono sample")
	// Full-scale sample clamps to +32767.
	assert.Equal(t, int16(32767), int16(binary.LittleEndian.Uint16(b[44+6:44+8])))
}

// S24-T02: an armed recorder retains audio samples, and the store exports them
// to a .wav, stamping the recording's AudioPath.
func TestRecorderCaptureAndStoreExport(t *testing.T) {
	r := NewRecorder(8000, &fakeClock{step: 1})
	r.Start("Bounce")
	r.RecordAudio(make([]float32, 4000)) // 0.5 s of audio
	rec, ok := r.Stop()
	require.True(t, ok)
	require.Len(t, rec.Samples, 4000, "samples retained for export")

	s := NewStore()
	saved := s.Add(rec)
	dir := t.TempDir()
	path, err := s.SaveWAV(saved.ID, dir)
	require.NoError(t, err)
	assert.FileExists(t, path)

	got, _ := s.Get(saved.ID)
	assert.Equal(t, path, got.AudioPath, "AudioPath recorded on the take")

	// A take with no audio can't be exported.
	empty := s.Add(Recording{Name: "midi only", SampleRate: 8000})
	_, err = s.SaveWAV(empty.ID, dir)
	assert.Error(t, err)
}
