package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S21-T04: the Live REC toggle arms and stops capture, and a stopped take shows
// up in the Recordings view list.
func TestRecordToggleAndRecordingsList(t *testing.T) {
	u, _ := newTestUI(t)

	// Build the toggle and the recordings view.
	toggle := u.buildRecordToggle()
	u.buildRecordingsView()

	assert.Equal(t, "● REC", toggle.Text)
	assert.False(t, u.ctrl.IsRecording())

	// Arm recording.
	u.toggleRecording()
	assert.True(t, u.ctrl.IsRecording())
	assert.Equal(t, "■ Stop · REC", toggle.Text)

	// Feed some MIDI so the take isn't empty, then stop.
	u.ctrl.VirtualNoteOn(60, 100)
	u.ctrl.VirtualNoteOff(60)
	u.toggleRecording()
	assert.False(t, u.ctrl.IsRecording())
	assert.Equal(t, "● REC", toggle.Text)

	// The finished take is listed.
	require.Len(t, u.recordings, 1)
	assert.Equal(t, "Take 1", u.recordings[0].Name)

	// Delete it via the transport action; the list empties.
	u.recSel = 0
	u.recAction("delete")
	assert.Empty(t, u.recordings)
}
