package provision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S22-T01: each model renders a valid systemd unit whose ExecStart launches the
// binary in that model's UI mode, and the install pair is written to disk.
func TestProvisionRendersPerModel(t *testing.T) {
	cases := map[string]string{
		"gs-49":      "-mode touch7",
		"gs-61":      "-mode touch7",
		"gs-desktop": "", // desktop default, no -mode
		"pizero":     "-mode pizero",
	}
	for id, wantMode := range cases {
		svc, err := RenderService(id, "/opt/gigsynth/gigsynth", "/opt/gigsynth/Live.sf2")
		require.NoError(t, err, id)
		assert.Contains(t, svc, "ExecStart=/opt/gigsynth/gigsynth", id)
		assert.Contains(t, svc, "-sf2 /opt/gigsynth/Live.sf2", id)
		assert.Contains(t, svc, "WantedBy=default.target", id)
		if wantMode != "" {
			assert.Contains(t, svc, wantMode, id)
		} else {
			assert.NotContains(t, svc, "-mode", id)
		}
	}

	// Install writes both artifacts.
	dir := t.TempDir()
	paths, err := Install("gs-61", "/opt/gigsynth/gigsynth", "", dir)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	for _, p := range paths {
		assert.FileExists(t, p)
	}
	ins, err := os.ReadFile(filepath.Join(dir, "install.sh"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(ins), "#!/usr/bin/env bash"), "install script is a bash script")
	assert.Contains(t, string(ins), "systemctl --user enable")
}

func TestProvisionUnknownModel(t *testing.T) {
	_, err := RenderService("nope", "bin", "")
	assert.Error(t, err)
	assert.Contains(t, ModelIDs(), "gs-49")
}
