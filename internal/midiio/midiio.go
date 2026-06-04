// Package midiio manages MIDI input devices (keyboards) and forwards parsed
// events to the application. It supports listening to several keyboards at once
// and tags every event with the source device name so the app can route each
// keyboard independently.
package midiio

import (
	"fmt"
	"sort"
	"sync"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

// Event is a normalized MIDI event tagged with its source device.
type Event struct {
	Device   string
	Type     EventType
	Channel  int
	Key      int // note number
	Velocity int
	Control  int // CC number
	Value    int // CC value / pitch-bend absolute (0..16383)
}

type EventType int

const (
	NoteOn EventType = iota
	NoteOff
	ControlChange
	PitchBend
)

// Manager owns the rtmidi driver and the set of open input ports.
type Manager struct {
	drv *rtmididrv.Driver

	mu      sync.Mutex
	open    map[string]openPort // device name -> open port + stop fn
	handler func(Event)
}

type openPort struct {
	in   drivers.In
	stop func()
}

// New creates the MIDI manager and initializes the rtmidi backend.
func New() (*Manager, error) {
	drv, err := rtmididrv.New()
	if err != nil {
		return nil, fmt.Errorf("midiio: init rtmidi: %w", err)
	}
	return &Manager{
		drv:  drv,
		open: make(map[string]openPort),
	}, nil
}

// SetHandler registers the callback invoked for every incoming event. It is
// called from a MIDI driver thread, so it must be cheap and non-blocking.
func (m *Manager) SetHandler(h func(Event)) {
	m.mu.Lock()
	m.handler = h
	m.mu.Unlock()
}

// ListInputs returns the names of available MIDI input devices.
func (m *Manager) ListInputs() ([]string, error) {
	ins, err := m.drv.Ins()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ins))
	for _, in := range ins {
		names = append(names, in.String())
	}
	sort.Strings(names)
	return names, nil
}

// Connected returns the names of currently open devices.
func (m *Manager) Connected() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, 0, len(m.open))
	for name := range m.open {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Connect opens a device by name and starts listening. Idempotent.
func (m *Manager) Connect(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.open[name]; ok {
		return nil
	}
	ins, err := m.drv.Ins()
	if err != nil {
		return err
	}
	var target drivers.In
	for _, in := range ins {
		if in.String() == name {
			target = in
			break
		}
	}
	if target == nil {
		return fmt.Errorf("midiio: input %q not found", name)
	}
	if err := target.Open(); err != nil {
		return fmt.Errorf("midiio: open %q: %w", name, err)
	}

	dev := name
	stop, err := target.Listen(func(data []byte, _ int32) {
		m.dispatch(dev, midi.Message(data))
	}, drivers.ListenConfig{})
	if err != nil {
		target.Close()
		return fmt.Errorf("midiio: listen %q: %w", name, err)
	}
	m.open[name] = openPort{in: target, stop: stop}
	return nil
}

// Disconnect stops listening and closes a device. Idempotent.
func (m *Manager) Disconnect(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.open[name]; ok {
		if p.stop != nil {
			p.stop()
		}
		p.in.Close()
		delete(m.open, name)
	}
}

// dispatch parses a raw MIDI message and forwards a normalized Event.
func (m *Manager) dispatch(dev string, msg midi.Message) {
	m.mu.Lock()
	h := m.handler
	m.mu.Unlock()
	if h == nil {
		return
	}

	var ch, key, vel, ctrl, val uint8
	var rel int16
	var abs uint16

	switch {
	case msg.GetNoteOn(&ch, &key, &vel):
		if vel == 0 { // running-status note-on with vel 0 == note off
			h(Event{Device: dev, Type: NoteOff, Channel: int(ch), Key: int(key)})
			return
		}
		h(Event{Device: dev, Type: NoteOn, Channel: int(ch), Key: int(key), Velocity: int(vel)})
	case msg.GetNoteOff(&ch, &key, &vel):
		h(Event{Device: dev, Type: NoteOff, Channel: int(ch), Key: int(key)})
	case msg.GetControlChange(&ch, &ctrl, &val):
		h(Event{Device: dev, Type: ControlChange, Channel: int(ch), Control: int(ctrl), Value: int(val)})
	case msg.GetPitchBend(&ch, &rel, &abs):
		h(Event{Device: dev, Type: PitchBend, Channel: int(ch), Value: int(abs)})
	}
}

// Close disconnects everything and shuts down the driver.
func (m *Manager) Close() {
	m.mu.Lock()
	for name, p := range m.open {
		if p.stop != nil {
			p.stop()
		}
		p.in.Close()
		delete(m.open, name)
	}
	m.mu.Unlock()
	if m.drv != nil {
		m.drv.Close()
	}
}
