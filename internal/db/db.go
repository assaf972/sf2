// Package db provides SQLite-backed persistence for GigSynth.
//
// Schema:
//
//	songs(id, name)
//	  └── parts(id, song_id, name, sort_order, master_gain)   -- many per song
//	        └── part_channels(part_id, channel 1..4, instrument + mix level)
//	midi_keyboards(id, manufacturer, model, midi_mapping)      -- supported-keyboard catalog
//
// Uses the pure-Go modernc.org/sqlite driver (no extra cgo), so the database
// layer never complicates cross-compilation.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// NumChannels is the fixed number of mix channels stored per part.
const NumChannels = 4

// Song is a piece of music; it has many ordered Parts.
type Song struct {
	ID   int64
	Name string
}

// Part is one section of a song (e.g. "Intro", "Chorus"). It captures the
// selected instrument and mix level for each of the 4 channels, plus the extra
// per-channel performance settings the live engine needs.
type Part struct {
	ID         int64
	SongID     int64
	Name       string
	SortOrder  int
	MasterGain float64
	Channels   [NumChannels]Channel
}

// Channel is the per-channel state inside a Part. Channel numbers are 1..4.
type Channel struct {
	Channel        int // 1..4
	Bank           int // selected instrument (SoundFont bank)
	Program        int // selected instrument (SoundFont program)
	InstrumentName string
	Level          int // mix level 0..127
	// Extra live-performance settings (kept so a Part fully restores the rig).
	Pan       int
	Mute      bool
	Enabled   bool
	Source    string
	KeyLow    int
	KeyHigh   int
	Transpose int
	// Per-keyboard effects (S02/S13/S14).
	ChorusOn      bool
	ChorusRate    float64 // Hz
	ChorusDepth   int     // %
	DelayOn       bool
	DelayTime     int // ms
	DelayFeedback int // %
	DelayMix      int // %
}

// Keyboard is an entry in the supported-MIDI-keyboard catalog.
type Keyboard struct {
	ID           int64
	Manufacturer string
	Model        string
	MidiMapping  string // free-form text (e.g. CC map notes)
}

// DB wraps the SQL handle.
type DB struct {
	sql *sql.DB
}

// DefaultPath returns the per-user database file path.
func DefaultPath() string {
	if cfg, err := os.UserConfigDir(); err == nil {
		return filepath.Join(cfg, "gigsynth", "gigsynth.db")
	}
	return "gigsynth.db"
}

// Open opens (creating if needed) the database at path and runs migrations.
func Open(path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	// Enable foreign-key enforcement (for ON DELETE CASCADE) per connection.
	dsn := path
	if !strings.Contains(dsn, "?") {
		dsn += "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	}
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Single writer connection keeps SQLite simple and avoids lock contention.
	sqlDB.SetMaxOpenConns(1)
	d := &DB{sql: sqlDB}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) migrate() error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`CREATE TABLE IF NOT EXISTS songs (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS parts (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id     INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
			name        TEXT NOT NULL,
			sort_order  INTEGER NOT NULL DEFAULT 0,
			master_gain REAL NOT NULL DEFAULT 0.6
		);`,
		`CREATE TABLE IF NOT EXISTS part_channels (
			part_id         INTEGER NOT NULL REFERENCES parts(id) ON DELETE CASCADE,
			channel         INTEGER NOT NULL,
			bank            INTEGER NOT NULL DEFAULT 0,
			program         INTEGER NOT NULL DEFAULT 0,
			instrument_name TEXT NOT NULL DEFAULT '',
			level           INTEGER NOT NULL DEFAULT 100,
			pan             INTEGER NOT NULL DEFAULT 64,
			mute            INTEGER NOT NULL DEFAULT 0,
			enabled         INTEGER NOT NULL DEFAULT 1,
			source          TEXT NOT NULL DEFAULT '',
			key_low         INTEGER NOT NULL DEFAULT 0,
			key_high        INTEGER NOT NULL DEFAULT 127,
			transpose       INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (part_id, channel)
		);`,
		`CREATE TABLE IF NOT EXISTS midi_keyboards (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			manufacturer TEXT NOT NULL,
			model        TEXT NOT NULL,
			midi_mapping TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS app_settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}
	for _, s := range stmts {
		if _, err := d.sql.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w (stmt: %.40s)", err, s)
		}
	}
	// Add per-keyboard effects columns to part_channels if an older database
	// predates them. SQLite has no "ADD COLUMN IF NOT EXISTS", so check first.
	fxCols := []struct{ name, decl string }{
		{"chorus_on", "INTEGER NOT NULL DEFAULT 0"},
		{"chorus_rate", "REAL NOT NULL DEFAULT 0.8"},
		{"chorus_depth", "INTEGER NOT NULL DEFAULT 50"},
		{"delay_on", "INTEGER NOT NULL DEFAULT 0"},
		{"delay_time", "INTEGER NOT NULL DEFAULT 300"},
		{"delay_feedback", "INTEGER NOT NULL DEFAULT 30"},
		{"delay_mix", "INTEGER NOT NULL DEFAULT 25"},
	}
	for _, c := range fxCols {
		if err := d.addColumnIfMissing("part_channels", c.name, c.decl); err != nil {
			return err
		}
	}
	return nil
}

// addColumnIfMissing adds a column to a table only if it isn't already present,
// making the FX migration safe to run against both fresh and existing databases.
func (d *DB) addColumnIfMissing(table, column, decl string) error {
	rows, err := d.sql.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil // already present
		}
	}
	_, err = d.sql.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, decl))
	return err
}
