# Guitar Rig — Audio DSP Stack & Library Choices

Source-of-truth for the guitar-rig audio backend (stories S26–S38). The design goal is a
**real-time guitar rig** — amp simulation + cabinet impulse-response + drive/modulation
pedals + reverb — that runs on macOS, Linux **and a Raspberry Pi 4/5 touchscreen**, with
a **permissive (non-GPL) license** so it ships in the commercial GigSynth product.

## "NAMM profiles" → NAM (Neural Amp Modeler)

The requested "NAMM profiles" are **NAM — Neural Amp Modeler** captures (file extension
`.nam`). NAM is the dominant open-source neural amp-capture format, the open analog to
Kemper/ToneX profiles. A `.nam` is JSON: architecture (WaveNet or LSTM) + config +
weights. It models the **amp/drive circuit only** — the speaker cabinet is a **separate
IR convolution** stage. Tone3000 is the public library of captures.

## Architecture

**Go orchestration + a single C/C++ DSP core behind one cgo boundary, processing one
block per audio callback** — mirroring the existing FluidSynth-via-cgo design. The
audio callback runs on a native thread; inside it we only call cgo `Process(float*, n)`
functions over pre-allocated buffers — **zero Go allocation, no channel ops, no logging**
— which keeps Go's GC entirely out of the audio path.

**Signal chain:**

```
guitar in → gate/comp → drive (fuzz/OD/dist) → NAM amp → cabinet IR → modulation/delay → reverb → out
```

Each stage is a plain `process(float* buf, int n)` C function exposed as an `fx.Node`.
Go owns the graph / presets / UI; C owns the samples.

## Recommended stack (all permissive)

| Layer | Choice | License | Wiring |
|---|---|---|---|
| Audio I/O | **malgo** (miniaudio), JACK backend optional on Pi | MIT / public-domain | cgo; native callback thread |
| Amp / drive capture | **NeuralAudio** (loads `.nam` WaveNet+LSTM + RTNeural) | **MIT** | cgo via its C API |
| Cabinet IR | **FFTConvolver + KissFFT** (partitioned / zero-latency) | MIT + BSD | cgo; `.wav` parsed in pure-Go `go-audio/wav` |
| Drive / mod / EQ / comp pedals | **FAUST → C**, vendored | generated code is yours | cgo; compiled into the DSP core |
| Reverb | **FAUST `dm.zita_rev1` + freeverb → C** | permissive | same as effects |

**Why these:**

- **NeuralAudio** (github.com/mikeoliphant/NeuralAudio, MIT) is the fastest NAM engine
  *and* the only one shipping a clean **C API** — ideal for a cgo shim. It loads NAM
  WaveNet/LSTM and RTNeural models. (The original `sdatkinson/NeuralAmpModelerCore` is
  also MIT but C++-only.)
- **FFTConvolver** (HiFi-LoFi, MIT) + **KissFFT** (BSD) give low-latency partitioned cab
  convolution with no GPL dependency. *Avoid* `zita-convolver` and `FFTW` — both GPL.
- **FAUST**: the compiler is GPL/LGPL **but the generated C code carries no such
  restriction** — proprietary use is explicitly permitted, and the DSP libs are
  LGPL-with-exception (freeverb origin BSD). FAUST gives a huge, battle-tested effect
  library that compiles to dependency-free C we cross-compile for the Pi.
- **malgo** (miniaudio, MIT/public-domain) is the best single cross-platform low-latency
  duplex I/O abstraction (CoreAudio / ALSA / JACK / WASAPI). Add a **JACK** backend for
  best latency on the Pi.

### Raspberry Pi feasibility (verified)

A full rig (NAM amp + IR cab + a few effects + reverb) is realistic on a **Pi 5** (below
~60% CPU for a Standard model @ 48 kHz / 256 frames; sub-10 ms latency with a good
interface) and **viable on a Pi 4** (Standard WaveNet runs real-time at a 96-sample
buffer, with headroom for an IR + light effects). The GPL reference project
`mikeoliphant/stompbox` proves the identical chain on a Pi 4 — study it as a blueprint,
**do not link it** (it's GPL-3).

## Licensing — must stay permissive

**Safe to ship closed:** NeuralAudio (MIT), NeuralAmpModelerCore (MIT), FFTConvolver
(MIT), KissFFT (BSD), pffft (BSD), FAUST-*generated* C (yours), miniaudio/malgo (PD/MIT),
PortAudio (MIT).

**Avoid linking:** Guitarix (GPL-2/3), stompbox / neural-amp-modeler-lv2 (GPL-3),
zita-convolver (GPL), FFTW (GPL; paid commercial license exists). LV2 *plugins* are
individually licensed — many GPL.

## Alternatives considered

- **Embed stompbox-core (GPL-3)** — fastest path to a working Pi rig, but forces the
  whole product to GPL-3. Only if GigSynth goes open-source.
- **Host LV2 plugins** (lilv) — max flexibility, but each plugin carries its own (often
  GPL) license and packaging on the Pi is heavier.

## Pro MP3 player extras (S37)

The Live Guitar Rig MP3 player adds **time-domain pitch shift** (resample-and-overlap or
WSOLA) and a **"remove guitar / vocal"** center-channel canceller (mid/side: subtract the
correlated center content, leaving the backing). Both are pure-Go block DSP and fully
unit-testable.

---

# Appendix — UI library decision (custom widgets in Fyne)

**Decision: stay on Fyne; build a custom skeuomorphic widget kit. Do _not_ migrate
frameworks.** (Stories S39–S40.)

The complaint "the UI looks nothing like the mockup" is real but **misattributed**: the
current UI is built **100% from stock Fyne widgets + the default theme** — there is not a
single custom `WidgetRenderer` or `canvas.*` primitive in the codebase. A theme swap
alone can't produce skeuomorphic knobs; a themed `widget.Slider` is still a slider.

Fyne **can** render the mockups via custom widgets: `canvas.Arc` (the knob value arc),
`Circle`, `Rectangle`, `LinearGradient`/`RadialGradient`, `Text`, and a `Raster` escape
hatch map almost one-to-one onto the SVG mockups. The plan: a `internal/ui/widgets`
package — `Knob`, `Slider`, `VUMeter`, `LEDToggle`, `ChannelStrip`, `Panel` — each a
`BaseWidget` + `WidgetRenderer`, plus a dark `fyne.Theme`. Expect ~80–90% fidelity.

| Option | Mockup fidelity | Pi touchscreen | Migration cost | Live-audio safety |
|---|---|---|---|---|
| **Fyne + custom widgets** (chosen) | Good (80–90%) | **Best** (already ships there) | **Lowest** (keep app, rebuild widgets) | Good |
| Gio | **Best** | Good (touch less proven on Pi) | High (full rewrite) | **Best** |
| Wails (web UI) | **Perfect** | **Worst** (WebKitGTK janky on Pi) | High (rewrite in web) | OK (IPC latency) |
| Ebiten | Good | OK (constant redraw burns CPU) | High + no app widgets | Weakest |

The two highest-fidelity options (Gio, Wails) each carry a Pi-specific risk on exactly
the device that has to be reliable on stage, plus a full UI rewrite. Fyne-with-custom-
widgets keeps the working `internal/app`/`engine`/`midiio`/`recording` plumbing and the
three existing shells (desktop / touch7 / pizero). Reconsider Gio only if a custom-widget
spike still disappoints *and* the desktop becomes the primary target.

## Sources

- NeuralAudio (MIT, C API) — github.com/mikeoliphant/NeuralAudio
- NeuralAmpModelerCore (MIT) — github.com/sdatkinson/NeuralAmpModelerCore
- stompbox (GPL-3 Pi rig reference) — github.com/mikeoliphant/stompbox
- NAM-on-Pi performance — blog.nostatic.org/2025/02/state-of-nam-on-raspberry-pi.html
- FFTConvolver (MIT) — github.com/HiFi-LoFi/FFTConvolver ; KissFFT (BSD)
- FAUST generated-code licensing — faustdoc.grame.fr/manual/faq/
- malgo (miniaudio) — github.com/gen2brain/malgo
- Fyne custom widgets — docs.fyne.io/extend/custom-widget/ ; Gio — gioui.org
