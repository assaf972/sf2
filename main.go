// GigSynth — a cross-platform FluidSynth wrapper with a unified mixer UI for
// gigging keyboard players. Layer/split up to 4 sounds from one SoundFont,
// driven by up to 3 MIDI keyboards, with recallable scenes per song.
package main

import (
	"flag"
	"log"
	"os"

	"gigsynth/internal/app"
	"gigsynth/internal/db"
	"gigsynth/internal/engine"
	"gigsynth/internal/midiio"
	"gigsynth/internal/ui"
)

func main() {
	sf := flag.String("sf2", "", "path to a SoundFont (.sf2) to load on startup")
	driver := flag.String("audio", "", "fluidsynth audio driver (empty = platform default)")
	mode := flag.String("mode", "", "ui mode: \"\" (desktop), touch7 (7-inch), pizero (small LCD)")
	flag.Parse()

	cfg := engine.DefaultConfig()
	if *driver != "" {
		cfg.AudioDriver = *driver
	}

	eng, err := engine.New(cfg)
	if err != nil {
		log.Fatalf("audio engine: %v", err)
	}
	defer eng.Close()

	m, err := midiio.New()
	if err != nil {
		log.Fatalf("midi: %v", err)
	}
	defer m.Close()

	store, err := app.NewStore(app.DefaultDir())
	if err != nil {
		log.Fatalf("scene store: %v", err)
	}

	database, err := db.Open(db.DefaultPath())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()
	lib := app.NewLibrary(database)

	ctrl := app.NewController(eng, m)
	ctrl.SetLibrary(lib)

	// Optional: load a SoundFont passed on the command line.
	if *sf != "" {
		if _, err := os.Stat(*sf); err == nil {
			if err := ctrl.LoadSoundFont(*sf); err != nil {
				log.Printf("warning: could not load %q: %v", *sf, err)
			}
		} else {
			log.Printf("warning: soundfont %q not found", *sf)
		}
	}

	ui.Run(ctrl, m, store, lib, *mode)
}
