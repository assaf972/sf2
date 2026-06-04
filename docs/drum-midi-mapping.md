# E-Drum MIDI Note Mapping Reference

This document is the source-of-truth for the **Drumming** feature (stories S32–S34).
It maps the MIDI notes that electronic drum kits send to the percussion voices in a
`drums.sf2` SoundFont. The encoded tables live in `internal/drummap`.

## Design summary

All five supported manufacturers ship a **General-MIDI-aligned default map** for the
core kit (kick, snare, hi-hat, toms, crash, ride). They diverge only on the *extra
zones* GM never defined — rim shots, cross-stick, cymbal bell/bow/edge, second crash,
hi-hat edge/half-open. Those zones are where per-manufacturer preset tables earn their
keep.

Two outliers to handle in code:

- **Roland** uses notes *below* the GM drum range for hi-hat edge (open edge `26`,
  closed edge `22`).
- **Yamaha DTX** swaps crash bow/edge versus Roland/GM **and differs between the
  5-pin MIDI-DIN port and USB-MIDI** on the same module.

Every kit lets the player reassign any pad, so these are *factory defaults* —
treat them as presets, not guarantees.

### Recommended encoding

A single canonical SF2 target map (GM-style) + per-manufacturer override tables for
the non-GM zones + **fold/alias rules** so an out-of-range incoming note never produces
silence:

- Canonical SF2 keys to sample: `35,36,37,38,40,41,42,43,44,45,46,47,48,49,50,51,52,53,55,57,58,59`.
- Roland HH edge `26→46`, `22→42`.
- Yamaha HH `78/79→46/42`, `83→46`; ride-edge `57↔52` normalized; crash bow/edge mapped
  by *semantic role*, not raw note.
- Alesis/Millenium HH splash `21→46`, half-open `23→46`.

Because **Roland, Alesis, Millenium and Donner** agree on the core + most zones, they can
share one "GM/Roland-style" preset; **Yamaha** needs its own preset (port-dependent +
swapped crash/ride).

## 1. General MIDI percussion key map (the baseline)

Channel 10, notes 35–81 (the playable e-kit range is essentially 35–59 plus a few outliers).

| Note | Instrument | Note | Instrument |
|---|---|---|---|
| 35 | Acoustic Bass Drum | 50 | High Tom |
| 36 | Bass Drum 1 | 51 | Ride Cymbal 1 |
| 37 | Side Stick / X-Stick | 52 | Chinese Cymbal |
| 38 | Acoustic Snare | 53 | Ride Bell |
| 39 | Hand Clap | 54 | Tambourine |
| 40 | Electric Snare | 55 | Splash Cymbal |
| 41 | Low Floor Tom | 56 | Cowbell |
| 42 | Closed Hi-Hat | 57 | Crash Cymbal 2 |
| 43 | High Floor Tom | 58 | Vibraslap |
| 44 | Pedal Hi-Hat | 59 | Ride Cymbal 2 |
| 45 | Low Tom | 60–81 | Latin perc (bongo, conga, timbale, agogo, …) |
| 46 | Open Hi-Hat | | |
| 47 | Low-Mid Tom | | |
| 48 | Hi-Mid Tom | | |
| 49 | Crash Cymbal 1 | | |

## 2. Roland TD / V-Drums (TD-17/25/27/50, VAD — the de-facto reference)

| Pad / Zone | Note | Pad / Zone | Note |
|---|---|---|---|
| Kick | 36 | HH Open (Bow) | 46 |
| Snare Head | 38 | HH Open (Edge) | **26** |
| Snare Rim | 40 | HH Closed (Bow) | 42 |
| Snare X-Stick | 37 | HH Closed (Edge) | **22** |
| Tom 1 Head / Rim | 48 / 50 | HH Pedal | 44 |
| Tom 2 Head / Rim | 45 / 47 | Crash 1 Bow / Edge | 49 / 55 |
| Tom 3 Head / Rim | 43 / 58 | Crash 2 Bow / Edge | 57 / 52 |
| Aux Head / Rim | 27 / 28 | Ride Bow / Edge / Bell | 51 / 59 / 53 |

Real hi-hat openness rides on **CC4 (Foot Controller)**; for SF2 mapping only the note matters.

## 3. Yamaha DTX (DTX502 representative; PRO/700/900 follow the scheme)

**Two columns** — the MIDI-DIN and USB-MIDI ports differ on the same module.

| Pad / Zone | DIN note | USB note |
|---|---|---|
| Kick 1 / Kick 2 | 36 / 35 | 36 / 35 |
| Snare Head / Open Rim / X-Stick | 38 / 40 / 37 | same |
| Tom 1 / 2 / 3 | 48 / 47 / 43 | same |
| Ride Bow / Edge / Cup | 51 / **52** / 53 | 51 / **57** / 53 |
| Crash Bow / Edge / Cup | **59** / **49** / 55 | 59 / 49 / 55 |
| HH Open / Open Edge | 46 / **78** | 46 / 46 |
| HH Closed / Closed Edge | 42 / **79** | 42 / 42 |
| HH Foot Close / Foot Splash | 44 / **83** | 44 / 46 |

Quirk: **crash bow = 59, crash edge = 49** — the opposite polarity from Roland/GM (which puts crash bow at 49).

## 4. Alesis (Nitro / Command / Crimson / Strike — shared house map)

Command/Crimson are identical; Nitro is a subset. Follows Roland closely.

| Pad / Zone | Note | Pad / Zone | Note |
|---|---|---|---|
| Kick | 36 | HH Closed / Open / Pedal | 42 / 46 / 44 |
| Snare Head / Rim | 38 / 40 | HH Splash | **21** |
| Tom 1 Head / Rim | 48 / 50 | Crash 1 Bow / Edge | 49 / 55 |
| Tom 2 Head / Rim | 45 / 47 | Crash 2 Bow / Edge | 57 / 52 |
| Tom 3 Head / Rim | 43 / 58 | Ride Bow / Edge / Bell | 51 / 59 / 53 |
| Tom 4 Head / Rim | 41 / 39 | HH Half-Open (Nitro) | **23** |

## 5. Millenium (Thomann MPS-150/600/750/850)

Pure GM for the 10 core triggers; larger modules add GM-aligned rim/crash2/ride zones
(rim 40, crash 2 57, ride edge 59, ride bell 53), all user-editable on the unit. Transmit channel fixed at 10.

| Pad | Note | Pad | Note |
|---|---|---|---|
| Kick | 36 | Crash 1 | 49 |
| Snare | 38 | Ride | 51 |
| Tom 1 / 2 / 3 | 48 / 45 / 43 | HH Closed / Open / Pedal | 42 / 46 / 44 |

## 6. Donner (DED-80 / DED-200 / DED-200X)

GM-aligned with Roland-style crash/ride zoning.

| Pad | Note | Pad | Note |
|---|---|---|---|
| Kick | 36 | Ride Bow / Edge | 51 / 59 |
| X-Stick (STICK) | 37 | Crash 1 Bow / Edge | 49 / 55 |
| Snare | 38 | Crash 2 Bow / Edge | 57 / 52 |
| Tom 1 / 2 / 3 | 48 / 45 / 43 | HH Closed / Open / Pedal | 42 / 46 / 44 |

DED-200 has crash/ride bow+edge only (no ride bell default — use 53 on DED-200X+).

## Sources

- General MIDI percussion map — en.wikipedia.org/wiki/General_MIDI
- Roland default note maps — support.roland.com (per-model "Default MIDI Note Map" articles)
- Yamaha DTX502 Reference Manual (note table, DIN vs USB columns) — usa.yamaha.com
- Alesis Nitro / Command / Crimson / Strike user guides — alesis.com
- Millenium MPS-150 / MPS-850 manuals — thomann.de / manualslib.com
- Donner DED-200/200X manual (MIDI NOTE table) — storage.donnermusic.com
