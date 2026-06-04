// Package provision renders the Raspberry Pi setup assets (a systemd user unit
// and an install script) for each GigSynth hardware model, so a build can boot
// straight into the rig. `gigsynth -provision <model>` writes them to disk.
//
// The live audio/display wiring (DAC overlay, DSI screen) is the operator's;
// this package produces the repeatable software setup that turns a flashed Pi
// into a GigSynth instrument.
package provision

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed assets/*
var assets embed.FS

// Model is one hardware target and the UI mode it launches.
type Model struct {
	ID     string
	Name   string
	Mode   string // gigsynth -mode value ("" = desktop)
	Screen string
}

// Models is the supported provisioning catalog, matching the product line.
var Models = map[string]Model{
	"gs-49":      {ID: "gs-49", Name: "GigSynth GS-49", Mode: "touch7", Screen: `5" DSI`},
	"gs-61":      {ID: "gs-61", Name: "GigSynth GS-61", Mode: "touch7", Screen: `7" DSI`},
	"gs-desktop": {ID: "gs-desktop", Name: "GigSynth GS-Desktop", Mode: "", Screen: `7" DSI`},
	"pizero":     {ID: "pizero", Name: "GigSynth Pi Zero kiosk", Mode: "pizero", Screen: "480x320 LCD"},
}

// ModelIDs returns the known model ids, sorted.
func ModelIDs() []string {
	ids := make([]string, 0, len(Models))
	for id := range Models {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ExecStart builds the launch command line for a model.
func ExecStart(m Model, bin, sf2 string) string {
	var b strings.Builder
	b.WriteString(bin)
	if m.Mode != "" {
		b.WriteString(" -mode " + m.Mode)
	}
	if sf2 != "" {
		b.WriteString(" -sf2 " + sf2)
	}
	return b.String()
}

// RenderService returns the systemd unit text for a model.
func RenderService(modelID, bin, sf2 string) (string, error) {
	m, ok := Models[modelID]
	if !ok {
		return "", fmt.Errorf("provision: unknown model %q (have %s)", modelID, strings.Join(ModelIDs(), ", "))
	}
	tmpl, err := assets.ReadFile("assets/gigsynth-kiosk.service")
	if err != nil {
		return "", err
	}
	s := string(tmpl)
	s = strings.ReplaceAll(s, "__DESC__", fmt.Sprintf("%s (%s)", m.Name, m.Screen))
	s = strings.ReplaceAll(s, "__EXECSTART__", ExecStart(m, bin, sf2))
	return s, nil
}

// RenderInstall returns the install script text for a model.
func RenderInstall(modelID, bin string) (string, error) {
	if _, ok := Models[modelID]; !ok {
		return "", fmt.Errorf("provision: unknown model %q", modelID)
	}
	tmpl, err := assets.ReadFile("assets/install.sh")
	if err != nil {
		return "", err
	}
	s := string(tmpl)
	s = strings.ReplaceAll(s, "__MODEL__", modelID)
	s = strings.ReplaceAll(s, "__BIN__", bin)
	return s, nil
}

// Install writes the rendered unit + install script for a model into outDir and
// returns the paths written.
func Install(modelID, bin, sf2, outDir string) ([]string, error) {
	svc, err := RenderService(modelID, bin, sf2)
	if err != nil {
		return nil, err
	}
	ins, err := RenderInstall(modelID, bin)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	svcPath := filepath.Join(outDir, "gigsynth-kiosk.service")
	insPath := filepath.Join(outDir, "install.sh")
	if err := os.WriteFile(svcPath, []byte(svc), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(insPath, []byte(ins), 0o755); err != nil {
		return nil, err
	}
	return []string{svcPath, insPath}, nil
}
