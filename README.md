# GigSynth

[![CI](https://github.com/assaf972/sf2/actions/workflows/ci.yml/badge.svg)](https://github.com/assaf972/sf2/actions/workflows/ci.yml)
[![CodeQL](https://github.com/assaf972/sf2/actions/workflows/codeql.yml/badge.svg)](https://github.com/assaf972/sf2/actions/workflows/codeql.yml)
[![Trivy](https://github.com/assaf972/sf2/actions/workflows/trivy.yml/badge.svg)](https://github.com/assaf972/sf2/actions/workflows/trivy.yml)

A cross-platform **FluidSynth wrapper with a unified mixer UI** for gigging
keyboard players. Load one SoundFont (`.sf2`/`.sf3`), layer or split up to **4
sounds**, drive them from up to **3 MIDI keyboards**, and recall per-song
**scenes** — on Windows, macOS, Linux, and Raspberry Pi.

## Stack & why

| Layer | Choice | Rationale |
|---|---|---|
| Language | **Go** | One self-contained binary per platform, trivial ARM cross-compile for the Pi, clean concurrency for several MIDI streams |
| Synth | **libfluidsynth via cgo** (`internal/engine`) | Direct C API (`fluid_synth_noteon`, per-channel CC7/CC10). FluidSynth synthesizes on its own real-time audio thread, so Go's GC never touches the audio path |
| MIDI in | **gomidi + rtmidi** (`internal/midiio`) | Per-device callbacks → independent routing of each keyboard. CoreMIDI / ALSA / WinMM |
| UI | **Fyne** (`internal/ui`) | Pure-Go, GPU-accelerated, single binary, **no webview/Node runtime** → robust on a Pi touchscreen at a gig |

```
main.go
└── internal/
    ├── engine/   cgo wrapper around libfluidsynth (audio + voices)
    ├── midiio/   MIDI device manager (per-keyboard events)
    ├── app/      domain model: Layers, Scenes, routing, persistence
    └── ui/       Fyne mixer, device panel, on-screen test keyboard
```

A **Layer** = one mixer channel (a SoundFont preset + volume/pan/mute + routing).
A **Scene** = a recallable set of 4 layers; this is the "preset" you pick per song.

Each keyboard has its own **four-effect insert chain** — Chorus → Phaser → Flanger →
Delay (`internal/fx`) — and you can **record** all MIDI + audio to replayable takes
(`internal/recording`, the Recordings view). There's also a hardware product line
(GS-49 / GS-61 / GS-Desktop) — see [docs/product-brochure.html](docs/product-brochure.html)
and the BOM in [docs/products/bom.md](docs/products/bom.md).

## Concepts

- **Layers / channels** — 4 strips, each on its own MIDI channel (0–3). Pick a
  sound, set volume/pan, mute, enable/disable. The first two are on by default
  (your "two channels"); enable all four when you want a bigger stack.
- **Routing per layer** — `From` selects which keyboard triggers the layer
  (or *All keyboards*); `Low`/`High` set a key-split range; `Transpose` shifts
  octaves. This gives you splits (bass below, pad above) and layers across
  multiple controllers.
- **Scenes** — save the whole rig with **Save As…**, recall instantly from the
  dropdown. Stored as JSON under your user config dir (`gigsynth/scenes`).
- **Effects** — per-keyboard Chorus, Phaser, Delay and Flanger inserts, each with its
  own On switch and knobs, stored in (and recalled with) every Part.
- **Recording** — a red **REC** toggle on the Live view captures all MIDI + audio; manage,
  replay, loop and delete takes in the **Recordings** view.
- **PANIC** — kills all stuck notes and re-applies the mixer.
- **Master** — global output gain.

## Build & run (macOS — this machine)

Dependencies (already installed here via Homebrew):

```sh
brew install go fluid-synth
```

Build and run:

```sh
go build -o gigsynth .
./gigsynth                 # then Load SoundFont… in the UI
./gigsynth -sf2 your.sf2   # or preload a SoundFont
```

A quick non-GUI audio self-test (`internal/engine` end-to-end):

```sh
go run ./cmd/smoke        # loads test.sf2, prints presets, plays a chord
```

## Linux / Raspberry Pi

Install the dev libraries, then build natively on the device (simplest, most
reliable for cgo):

```sh
# Debian/Ubuntu/Raspberry Pi OS
sudo apt install golang libfluidsynth-dev librtmidi-dev libasound2-dev \
                 libgl1-mesa-dev xorg-dev pkg-config
go build -o gigsynth .
```

Pi latency tips: use a USB audio interface, run with PipeWire/ALSA, and keep the
low-latency defaults (`-audio alsa`, period 64×2 @ 48 kHz in `engine.DefaultConfig`).

## Windows

Install MSYS2 + the FluidSynth and rtmidi dev packages (or vcpkg), ensure a C
toolchain is on `PATH` (cgo requires it), then `go build`. The default audio
driver resolves to WASAPI/DirectSound automatically.

## Cross-compiling

Because of cgo (FluidSynth + rtmidi are C), the cleanest path is to **build on
each target OS/arch**. For the Pi, building on the device or in an ARM
container/VM avoids cross-cgo toolchain pain. Pure `GOOS/GOARCH` cross-builds
won't work without a matching C cross-toolchain and target libraries.

## Latency tuning

`engine.DefaultConfig()` uses 48 kHz, period size 64, 2 periods — a good live
default. Lower the period size for less latency (more CPU / xrun risk); raise it
if you hear dropouts.
# sf2
