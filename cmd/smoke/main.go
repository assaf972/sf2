package main

import (
	"fmt"
	"os"
	"time"

	"gigsynth/internal/engine"
)

func main() {
	cfg := engine.DefaultConfig()
	eng, err := engine.New(cfg)
	if err != nil {
		fmt.Println("ENGINE INIT FAILED:", err)
		os.Exit(1)
	}
	defer eng.Close()
	fmt.Println("engine + audio driver OK")

	presets, err := eng.LoadSoundFont("test.sf2")
	if err != nil {
		fmt.Println("LOAD FAILED:", err)
		os.Exit(1)
	}
	fmt.Printf("loaded soundfont: %d presets\n", len(presets))
	for i, p := range presets {
		if i >= 5 {
			break
		}
		fmt.Printf("  bank=%d prog=%d %q\n", p.Bank, p.Program, p.Name)
	}
	// Play a chord on channel 0 for ~1s to exercise the audio path.
	_ = eng.SelectProgram(0, presets[0].Bank, presets[0].Program)
	eng.SetChannelVolume(0, 110)
	for _, n := range []int{60, 64, 67} {
		eng.NoteOn(0, n, 100)
	}
	time.Sleep(1200 * time.Millisecond)
	for _, n := range []int{60, 64, 67} {
		eng.NoteOff(0, n)
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Println("played test chord (channel 0)")
}
