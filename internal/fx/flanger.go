package fx

import "math"

// Flanger is a short LFO-modulated delay with feedback and four controls:
//
//	"rate"     LFO speed in Hz
//	"depth"    sweep amount, percent 0..100
//	"feedback" regeneration, percent 0..100
//	"mix"      wet level, percent 0..100
//
// The very short (sub-15 ms) swept delay summed with the dry signal produces the
// classic flanger "jet" comb sweep. When disabled it passes the dry signal
// through unchanged. Real-time-safe: no allocation in Process.
type Flanger struct {
	sr       float64
	enabled  bool
	rate     float64 // Hz
	depth    float64 // 0..1
	feedback float64 // 0..1
	mix      float64 // 0..1

	buf   []float32
	w     int
	phase float64
}

// NewFlanger returns a flanger at the given sample rate (15 ms delay buffer).
func NewFlanger(sampleRate float64) *Flanger {
	return &Flanger{
		sr:       sampleRate,
		rate:     0.25,
		depth:    0.7,
		feedback: 0.4,
		mix:      0.5,
		buf:      make([]float32, int(sampleRate*0.015)+2),
	}
}

func (f *Flanger) SetEnabled(on bool) { f.enabled = on }
func (f *Flanger) Enabled() bool       { return f.enabled }

func (f *Flanger) SetParam(name string, v float64) {
	switch name {
	case "rate":
		f.rate = v
	case "depth":
		f.depth = clamp01(v / 100)
	case "feedback":
		f.feedback = clamp01(v / 100)
	case "mix":
		f.mix = clamp01(v / 100)
	}
}

func (f *Flanger) Process(buf []float32) {
	if !f.enabled {
		return
	}
	n := len(f.buf)
	base := 0.001 * f.sr          // 1 ms base delay
	swing := f.depth * 0.009 * f.sr // up to 9 ms sweep
	twoPiRateOverSr := 2 * math.Pi * f.rate / f.sr

	for i := range buf {
		f.phase += twoPiRateOverSr
		if f.phase > 2*math.Pi {
			f.phase -= 2 * math.Pi
		}
		delay := base + swing*math.Sin(f.phase)
		if delay < 0 {
			delay = 0
		}

		// Linear-interpolated read `delay` samples behind the write head.
		rd := float64(f.w) - delay
		for rd < 0 {
			rd += float64(n)
		}
		i0 := int(rd)
		frac := rd - float64(i0)
		i1 := (i0 + 1) % n
		wet := f.buf[i0%n]*float32(1-frac) + f.buf[i1]*float32(frac)

		dry := buf[i]
		f.buf[f.w] = dry + wet*float32(f.feedback)
		f.w = (f.w + 1) % n
		buf[i] = dry + wet*float32(f.mix)
	}
}
