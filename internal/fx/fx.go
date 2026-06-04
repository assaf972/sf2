// Package fx provides the per-keyboard effects insert chain.
//
// FluidSynth's built-in chorus/reverb are global and it has no delay, so to give
// each keyboard its OWN chorus and delay GigSynth renders these as independent
// Go DSP nodes that process that keyboard's audio buffer. Each keyboard owns one
// Chain (Chorus -> Delay); nodes are real-time-safe (no allocation in Process)
// and bypass to the dry signal when disabled.
package fx

// Node is one effect in a keyboard's insert chain.
type Node interface {
	// Process transforms a mono audio buffer in place.
	Process(buf []float32)
	SetEnabled(on bool)
	Enabled() bool
	// SetParam sets a named parameter in human units (see each node's doc).
	SetParam(name string, v float64)
}

// Chain is the per-keyboard insert chain: Chorus -> Phaser -> Flanger -> Delay.
// Modulation effects sit before the delay so their movement is echoed.
type Chain struct {
	Chorus  *Chorus
	Phaser  *Phaser
	Flanger *Flanger
	Delay   *Delay
}

// NewChain builds a disabled-by-default chain at the given sample rate.
func NewChain(sampleRate float64) *Chain {
	return &Chain{
		Chorus:  NewChorus(sampleRate),
		Phaser:  NewPhaser(sampleRate),
		Flanger: NewFlanger(sampleRate),
		Delay:   NewDelay(sampleRate),
	}
}

// Process runs the buffer through chorus, phaser, flanger, then delay.
func (c *Chain) Process(buf []float32) {
	c.Chorus.Process(buf)
	c.Phaser.Process(buf)
	c.Flanger.Process(buf)
	c.Delay.Process(buf)
}

var (
	_ Node = (*Chorus)(nil)
	_ Node = (*Phaser)(nil)
	_ Node = (*Flanger)(nil)
	_ Node = (*Delay)(nil)
)
