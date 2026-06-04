package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S18-T01: PANIC silences everything first (engine reset), then re-applies the
// mixer (CC7 volume / CC10 pan) so levels survive the reset.
func TestPanicResetsThenReappliesMixer(t *testing.T) {
	f := &FakeEngine{}
	c := NewController(f, nil)
	c.SetScene(Scene{
		Name:       "t",
		MasterGain: 0.6,
		Layers: []Layer{
			{Channel: 0, Enabled: true, Volume: 100, Pan: 64},
			{Channel: 1, Enabled: true, Volume: 80, Pan: 30},
		},
	})
	f.Reset()

	c.Panic()

	require.NotEmpty(t, f.Calls)
	assert.Equal(t, "Panic", f.Calls[0], "PANIC must silence before re-applying")

	// After the reset, the mixer CCs are re-sent for both channels.
	idxPanic := 0
	idxVol0 := indexOf(f.Calls, "CC ch=0 ctrl=7 val=100")
	idxPan0 := indexOf(f.Calls, "CC ch=0 ctrl=10 val=64")
	idxVol1 := indexOf(f.Calls, "CC ch=1 ctrl=7 val=80")
	assert.Greater(t, idxVol0, idxPanic, "volume re-applied after reset")
	assert.Greater(t, idxPan0, idxPanic, "pan re-applied after reset")
	assert.Greater(t, idxVol1, idxPanic, "channel 1 volume re-applied after reset")
}

func indexOf(s []string, want string) int {
	for i, v := range s {
		if v == want {
			return i
		}
	}
	return -1
}
