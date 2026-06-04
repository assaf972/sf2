package recording

import (
	"fmt"
	"path/filepath"
	"sync"
)

// Store is the in-memory catalog of finished takes shown on the Recordings
// screen. It assigns stable IDs and supports the page's actions: list, get,
// delete and "delete all". Safe for concurrent use.
type Store struct {
	mu   sync.Mutex
	recs []Recording
	seq  int
}

// NewStore returns an empty store.
func NewStore() *Store { return &Store{} }

// Add stores a recording, assigning an ID if it has none, and returns the stored
// copy (with ID populated).
func (s *Store) Add(r Recording) Recording {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		s.seq++
		r.ID = formatID(s.seq)
	}
	s.recs = append(s.recs, r)
	return r
}

// List returns all recordings, newest first.
func (s *Store) List() []Recording {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Recording, len(s.recs))
	for i, r := range s.recs {
		out[len(s.recs)-1-i] = r
	}
	return out
}

// Get returns a recording by ID.
func (s *Store) Get(id string) (Recording, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.recs {
		if r.ID == id {
			return r, true
		}
	}
	return Recording{}, false
}

// Delete removes a recording by ID, reporting whether it existed.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.recs {
		if r.ID == id {
			s.recs = append(s.recs[:i], s.recs[i+1:]...)
			return true
		}
	}
	return false
}

// DeleteAll clears every recording ("Delete all" on the Recordings page).
func (s *Store) DeleteAll() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.recs)
	s.recs = nil
	return n
}

// Len reports how many recordings are stored.
func (s *Store) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.recs)
}

// SaveWAV writes the captured audio of recording id to <dir>/<id>.wav, records
// the path on the stored recording, and returns it. Errors if the take has no
// audio samples or the id is unknown.
func (s *Store) SaveWAV(id, dir string) (string, error) {
	s.mu.Lock()
	idx := -1
	for i, r := range s.recs {
		if r.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		s.mu.Unlock()
		return "", fmt.Errorf("recording: unknown id %q", id)
	}
	rec := s.recs[idx]
	s.mu.Unlock()

	if len(rec.Samples) == 0 {
		return "", fmt.Errorf("recording %q has no captured audio", id)
	}
	path := filepath.Join(dir, id+".wav")
	if err := WriteWAV(path, int(rec.SampleRate), rec.Samples); err != nil {
		return "", err
	}

	s.mu.Lock()
	if idx < len(s.recs) && s.recs[idx].ID == id {
		s.recs[idx].AudioPath = path
	}
	s.mu.Unlock()
	return path, nil
}
