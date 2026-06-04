# GigSynth Product Line — Bill of Materials & Pricing (Brochure Mockup)

This document estimates the bill of materials (BOM), all-in unit cost, and suggested
pricing for the three GigSynth products: the **GS-49** keyboard controller, the
**GS-61** keyboard controller, and the **GS-Desktop** sound-module-style unit. It is
intended as a **marketing brochure mockup** to size up the product line and sanity-check
margins. All figures are rough engineering estimates for a small production batch.

> **⚠️ DISCLAIMER — READ BEFORE USING ANY NUMBER**
> **Every price in this document is an OFFLINE ESTIMATE based on 2026 knowledge of the
> hardware market. NO LIVE SOURCES WERE FETCHED — no vendor pages, distributor feeds,
> or web prices were checked.** All prices are marked **"(est.)"** for this reason.
> Currency is USD unless noted. These numbers are NOT web-verified and may be stale,
> regional, or wrong. **Verify every line item against live supplier quotes before
> issuing any real quote, purchase order, or published price.**

---

## Reference donor controllers — estimated market prices

These are the off-the-shelf controllers we reference as donor units or price anchors.

| Donor / reference | Type | Street price (USD, est.) | Notes |
|---|---|---|---|
| Nektar Impact GX49 | 49-key USB MIDI controller | $99 (est.) | Price anchor / donor for **GS-49** |
| Generic 61-key USB MIDI controller | 61-key USB MIDI (AliExpress-class) | $100 (est.) | ~ILS 363 (est.); donor for **GS-61** |
| Generic semi-weighted keybed module | 49/61-key keybed + chassis only | $55–90 (est.) | Alternative to a full donor unit |
| Official Raspberry Pi 7" DSI touchscreen | 7" capacitive DSI display | $60–100 (est.) | Used in GS-61 / GS-Desktop |
| Generic 5" DSI touchscreen | 5" capacitive DSI display | $40–60 (est.) | Used in GS-49 |

---

## GS-49 — 49-key controller (5" touch, Pi 5 4GB)

49 keys, 4 knobs + 4 sliders, embedded Raspberry Pi 5 + 5" DSI touchscreen,
stereo line out + headphone out.

| Component | Part / example | Unit cost (USD, est.) | Notes |
|---|---|---|---|
| Donor controller | Nektar Impact GX49-class | $99 (est.) | Keybed, case, pads/wheels harvested |
| SBC | Raspberry Pi 5 (4GB) | $60 (est.) | 4GB sufficient for synth engine |
| Display | 5" DSI capacitive touchscreen | $50 (est.) | DSI ribbon to Pi |
| Audio DAC/HAT | HiFiBerry DAC2 | $45 (est.) | Stereo line out + HP out |
| USB MIDI host | USB host cable / adapter | $8 (est.) | Reads donor MIDI over USB |
| Storage | microSD 32GB (A1) | $8 (est.) | OS + soundfonts |
| Cooling | Pi 5 active cooler | $10 (est.) | Sustained DSP load |
| Power | 5V/5A USB-C PSU | $14 (est.) | Official-class supply |
| Controls | 4 knobs + 4 sliders + encoders | $18 (est.) | Pots/encoders + caps |
| Wiring | Custom wiring harness | $12 (est.) | Internal looms, connectors |
| Enclosure | 3D-printed mounts / panel inserts | $30 (est.) | Reuses donor shell + printed parts |
| Misc | Cabling, fasteners, standoffs | $10 (est.) | Consumables |
| **Hardware subtotal** | | **$364 (est.)** | Sum of above |

---

## GS-61 — 61-key controller (7" touch, Pi 5 8GB)

61 keys, 8 knobs + 8 sliders, embedded Raspberry Pi 5 + 7" DSI touchscreen,
stereo out + aux out.

| Component | Part / example | Unit cost (USD, est.) | Notes |
|---|---|---|---|
| Donor controller | Generic 61-key USB MIDI controller | $100 (est.) | ~ILS 363 (est.); keybed + chassis |
| SBC | Raspberry Pi 5 (8GB) | $80 (est.) | 8GB for larger soundfonts |
| Display | Official 7" DSI touchscreen | $85 (est.) | Larger UI surface |
| Audio DAC/HAT | HiFiBerry DAC2 | $45 (est.) | Stereo out + aux out |
| MIDI I/O | USB MIDI host + DIN I/O board | $12 (est.) | USB host + DIN expansion |
| Storage | microSD 32GB (A1) | $8 (est.) | OS + soundfonts |
| Cooling | Pi 5 active cooler | $10 (est.) | Sustained DSP load |
| Power | 5V/5A USB-C PSU | $14 (est.) | Official-class supply |
| Controls | 8 knobs + 8 sliders + encoders | $32 (est.) | Double the GS-49 control set |
| Wiring | Custom wiring harness | $16 (est.) | Longer looms, more controls |
| Enclosure | Sheet-metal / custom panel | $45 (est.) | Larger chassis, metal top |
| Misc | Cabling, fasteners, standoffs | $12 (est.) | Consumables |
| **Hardware subtotal** | | **$459 (est.)** | Sum of above |

---

## GS-Desktop — sound module (no keys, 7" touch, Pi 5 8GB)

Roland-sound-module-style desktop unit (no keybed): 7" touchscreen, 4 knobs + 4 sliders,
stereo out + aux out, inputs for up to **three** external MIDI keyboards (USB host + DIN MIDI in).

| Component | Part / example | Unit cost (USD, est.) | Notes |
|---|---|---|---|
| SBC | Raspberry Pi 5 (8GB) | $80 (est.) | 8GB for multi-keyboard layering |
| Audio + MIDI HAT | pisound | $120 (est.) | DAC + **DIN MIDI in/out** built in |
| Display | Official 7" DSI touchscreen | $85 (est.) | Front-panel UI |
| USB MIDI host | 3-port powered USB hub | $15 (est.) | Hosts up to 3 USB MIDI keyboards |
| DIN MIDI in | (via pisound) | $0 (est.) | Provided by pisound HAT |
| Storage | microSD 32GB (A1) | $8 (est.) | OS + soundfonts |
| Cooling | Pi 5 active cooler | $10 (est.) | Sustained DSP load |
| Power | 5V/5A USB-C PSU | $14 (est.) | Powers Pi + hub |
| Controls | 4 knobs + 4 sliders + encoders | $18 (est.) | Front-panel controls |
| Wiring | Custom wiring harness | $14 (est.) | Internal looms, jacks |
| Enclosure | Sheet-metal desktop chassis | $50 (est.) | Module-style metal box |
| Misc | Cabling, fasteners, standoffs | $12 (est.) | Consumables |
| **Hardware subtotal** | | **$426 (est.)** | Sum of above |

---

## Assembly, test & packaging

Allow roughly **$25–60 per unit** for assembly, flashing/test, QC, and retail packaging,
scaled to product complexity. Applied per product below:

| Product | Assembly / test / packaging (USD, est.) |
|---|---|
| GS-49 | $30 (est.) |
| GS-61 | $45 (est.) |
| GS-Desktop | $40 (est.) |

---

## Margin model

All-in unit cost = hardware subtotal + assembly/test/packaging. MSRP set at ~2.2–2.6×
all-in, tiered so GS-49 is cheapest, GS-Desktop mid, and GS-61 top (largest screen + 61 keys).

| Product | Hardware subtotal (est.) | Assembly (est.) | All-in unit cost (est.) | Kit / DIY price (est.) | Multiplier | Assembled MSRP (est.) |
|---|---|---|---|---|---|---|
| GS-49 | $364 | $30 | **$394** | **$320** | 2.3× | **$899** |
| GS-Desktop | $426 | $40 | **$466** | **$380** | 2.5× | **$1,149** |
| GS-61 | $459 | $45 | **$504** | **$410** | 2.4× | **$1,199** |

*MSRP math: GS-49 $394 × 2.3 ≈ $906 → $899; GS-Desktop $466 × 2.5 = $1,165 → $1,149;
GS-61 $504 × 2.4 ≈ $1,210 → $1,199 (rounded to clean retail price points).*

---

## Notes / assumptions

- **No live sources.** All prices are 2026 offline estimates, marked "(est.)", and must
  be verified against live supplier quotes before any real quote or published price.
- **Volume:** costs assume a **small batch of ~100 units**; single-unit prototype costs
  would be higher, and 1,000+ volume would lower donor/SBC/display unit costs.
- **Labor rate assumption:** assembly/test figures assume in-house build at roughly
  $25–35/hr fully loaded, ~1–2 hours per unit depending on model.
- **Donor strategy:** GS-49 and GS-61 reuse a complete donor controller (keybed, chassis,
  wheels). A keybed-only sourcing path could trim cost but adds chassis tooling.
- **DAC/MIDI choice:** GS-49/GS-61 use HiFiBerry DAC2 (audio only) plus a separate USB/DIN
  MIDI path; GS-Desktop uses **pisound**, which costs more (~$120 est.) but bundles DIN MIDI.
- **Kit / DIY price** covers hardware + a build guide, no assembly labor or warranty;
  **Assembled MSRP** includes assembly, test, packaging, margin, and support headroom.
- Prices exclude shipping, duties/VAT, payment processing fees, and warranty reserve.
- Currency conversions (e.g., ILS) are approximate and for reference only.
