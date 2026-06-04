package fx

import "math"

// Phaser is a 4-stage all-pass phaser with three controls:
//
//	"rate"     LFO speed in Hz
//	"depth"    sweep amount, percent 0..100
//	"feedback" resonance / regeneration, percent 0..100
//
// It sweeps a cascade of first-order all-pass filters with a low-frequency
// oscillator, producing the classic moving notches. When disabled it passes the
// dry signal through unchanged. Real-time-safe: no allocation in Process.
type Phaser struct {
	sr       float64
	enabled  bool
	rate     float64 // Hz
	depth    float64 // 0..1
	feedback float64 // 0..1

	phase float64
	zm    [phaserStages]float64 // per-stage all-pass memory
	fb    float64               // last output, fed back
}

const phaserStages = 4

// NewPhaser returns a phaser at the given sample rate with musical defaults.
func NewPhaser(sampleRate float64) *Phaser {
	return &Phaser{
		sr:       sampleRate,
		rate:     0.5,
		depth:    0.6,
		feedback: 0.3,
	}
}

func (p *Phaser) SetEnabled(on bool) { p.enabled = on }
func (p *Phaser) Enabled() bool       { return p.enabled }

func (p *Phaser) SetParam(name string, v float64) {
	switch name {
	case "rate":
		p.rate = v
	case "depth":
		p.depth = clamp01(v / 100)
	case "feedback":
		p.feedback = clamp01(v / 100)
	}
}

func (p *Phaser) Process(buf []float32) {
	if !p.enabled {
		return
	}
	twoPiRateOverSr := 2 * math.Pi * p.rate / p.sr
	// Sweep the all-pass corner between ~200 Hz and ~1600 Hz, scaled by depth.
	const fMin, fMax = 200.0, 1600.0

	for i := range buf {
		p.phase += twoPiRateOverSr
		if p.phase > 2*math.Pi {
			p.phase -= 2 * math.Pi
		}
		lfo := 0.5 * (1 + math.Sin(p.phase)) // 0..1
		fc := fMin + (fMax-fMin)*p.depth*lfo
		// First-order all-pass coefficient for corner fc.
		t := math.Tan(math.Pi * fc / p.sr)
		a := (t - 1) / (t + 1)

		dry := float64(buf[i])
		x := dry + p.fb*p.feedback
		// Cascade of identical first-order all-pass sections.
		for s := 0; s < phaserStages; s++ {
			y := a*x + p.zm[s]
			p.zm[s] = x - a*y
			x = y
		}
		p.fb = x
		buf[i] = float32(0.5*dry + 0.5*x)
	}
}
