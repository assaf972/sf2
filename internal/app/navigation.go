package app

import "gigsynth/internal/db"

// Setlist navigation: the Controller tracks the current song's ordered parts and
// a cursor, so the Live view's Prev/Next Part shuttle (and the 61-key controller
// in Pi Zero mode) can recall whole parts hands-free.

// SetLibrary attaches the song/part store used for part recall.
func (c *Controller) SetLibrary(l *Library) {
	c.mu.Lock()
	c.lib = l
	c.mu.Unlock()
}

// LoadSong loads a song's parts and recalls the first one.
func (c *Controller) LoadSong(songID int64, songName string) error {
	c.mu.Lock()
	lib := c.lib
	c.mu.Unlock()
	if lib == nil {
		return errNoLibrary
	}
	parts, err := lib.DB().ListParts(songID)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.songName = songName
	c.parts = parts
	c.partIdx = 0
	c.mu.Unlock()
	c.recallCurrent()
	return nil
}

// PartCount / CurrentPartIndex / CurrentPartName / CurrentSong expose cursor state.
func (c *Controller) PartCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.parts)
}
func (c *Controller) CurrentPartIndex() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.partIdx
}
func (c *Controller) CurrentPartName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.partIdx < 0 || c.partIdx >= len(c.parts) {
		return ""
	}
	return c.parts[c.partIdx].Name
}
func (c *Controller) CurrentSong() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.songName
}

// NextPart / PrevPart move the cursor (clamped) and recall that part.
func (c *Controller) NextPart() { c.movePart(+1) }
func (c *Controller) PrevPart() { c.movePart(-1) }

// GoToPart jumps to an absolute part index (clamped) and recalls it.
func (c *Controller) GoToPart(i int) {
	c.mu.Lock()
	if len(c.parts) == 0 {
		c.mu.Unlock()
		return
	}
	c.partIdx = clamp(i, 0, len(c.parts)-1)
	c.mu.Unlock()
	c.recallCurrent()
}

func (c *Controller) movePart(delta int) {
	c.mu.Lock()
	if len(c.parts) == 0 {
		c.mu.Unlock()
		return
	}
	c.partIdx = clamp(c.partIdx+delta, 0, len(c.parts)-1)
	c.mu.Unlock()
	c.recallCurrent()
}

// recallCurrent applies the current part as the live scene (SetScene issues a
// PANIC first, so notes never stick across a part change).
func (c *Controller) recallCurrent() {
	c.mu.RLock()
	var p db.Part
	ok := c.partIdx >= 0 && c.partIdx < len(c.parts)
	if ok {
		p = c.parts[c.partIdx]
	}
	c.mu.RUnlock()
	if ok {
		c.SetScene(sceneFromPart(p))
	}
}
