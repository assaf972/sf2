package fx

// Delay is a feedback delay line with three controls:
//
//	"time"     echo spacing in milliseconds
//	"feedback" repeat amount, percent 0..100
//	"mix"      wet level, percent 0..100
//
// When disabled it passes the dry signal through unchanged.
type Delay struct {
	sr       float64
	enabled  bool
	timeMs   float64
	feedback float64 // 0..1
	mix      float64 // 0..1
	buf      []float32
	w        int
}

// NewDelay returns a delay sized for up to ~2s at the given sample rate.
func NewDelay(sampleRate float64) *Delay {
	return &Delay{
		sr:       sampleRate,
		timeMs:   300,
		feedback: 0.30,
		mix:      0.25,
		buf:      make([]float32, int(sampleRate*2)+1),
	}
}

func (d *Delay) SetEnabled(on bool) { d.enabled = on }
func (d *Delay) Enabled() bool       { return d.enabled }

func (d *Delay) SetParam(name string, v float64) {
	switch name {
	case "time":
		d.timeMs = v
	case "feedback":
		d.feedback = clamp01(v / 100)
	case "mix":
		d.mix = clamp01(v / 100)
	}
}

func (d *Delay) delaySamples() int {
	n := int(d.timeMs / 1000 * d.sr)
	if n < 1 {
		n = 1
	}
	if n >= len(d.buf) {
		n = len(d.buf) - 1
	}
	return n
}

func (d *Delay) Process(buf []float32) {
	if !d.enabled {
		return
	}
	ds := d.delaySamples()
	n := len(d.buf)
	for i := range buf {
		r := (d.w - ds + n) % n
		delayed := d.buf[r]
		d.buf[d.w] = buf[i] + delayed*float32(d.feedback)
		buf[i] = buf[i] + delayed*float32(d.mix)
		d.w = (d.w + 1) % n
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
