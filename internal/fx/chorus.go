package fx

import "math"

// Chorus is an LFO-modulated delay with two controls:
//
//	"rate"  LFO speed in Hz
//	"depth" modulation amount, percent 0..100
//
// When disabled it passes the dry signal through unchanged.
type Chorus struct {
	sr      float64
	enabled bool
	rate    float64 // Hz
	depth   float64 // 0..1
	buf     []float32
	w       int
	phase   float64
}

// NewChorus returns a chorus at the given sample rate (50 ms delay buffer).
func NewChorus(sampleRate float64) *Chorus {
	return &Chorus{
		sr:    sampleRate,
		rate:  0.8,
		depth: 0.5,
		buf:   make([]float32, int(sampleRate*0.05)+2),
	}
}

func (c *Chorus) SetEnabled(on bool) { c.enabled = on }
func (c *Chorus) Enabled() bool       { return c.enabled }

func (c *Chorus) SetParam(name string, v float64) {
	switch name {
	case "rate":
		c.rate = v
	case "depth":
		c.depth = clamp01(v / 100)
	}
}

func (c *Chorus) Process(buf []float32) {
	if !c.enabled {
		return
	}
	n := len(c.buf)
	base := 0.015 * c.sr          // 15 ms base delay
	swing := c.depth * 0.010 * c.sr // up to 10 ms modulation
	twoPiRateOverSr := 2 * math.Pi * c.rate / c.sr

	for i := range buf {
		c.phase += twoPiRateOverSr
		if c.phase > 2*math.Pi {
			c.phase -= 2 * math.Pi
		}
		delay := base + swing*math.Sin(c.phase)

		// Linear-interpolated read `delay` samples behind the write head.
		rd := float64(c.w) - delay
		for rd < 0 {
			rd += float64(n)
		}
		i0 := int(rd)
		frac := rd - float64(i0)
		i1 := (i0 + 1) % n
		wet := c.buf[i0%n]*float32(1-frac) + c.buf[i1]*float32(frac)

		dry := buf[i]
		c.buf[c.w] = dry
		c.w = (c.w + 1) % n
		buf[i] = 0.7*dry + 0.7*wet
	}
}
