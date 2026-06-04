package db

import "database/sql"

// ---- songs ----

// CreateSong inserts a song and returns it with its new ID.
func (d *DB) CreateSong(name string) (Song, error) {
	res, err := d.sql.Exec("INSERT INTO songs(name) VALUES(?)", name)
	if err != nil {
		return Song{}, err
	}
	id, _ := res.LastInsertId()
	return Song{ID: id, Name: name}, nil
}

// ListSongs returns all songs ordered by name.
func (d *DB) ListSongs() ([]Song, error) {
	rows, err := d.sql.Query("SELECT id, name FROM songs ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Song
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// RenameSong updates a song's name.
func (d *DB) RenameSong(id int64, name string) error {
	_, err := d.sql.Exec("UPDATE songs SET name=? WHERE id=?", name, id)
	return err
}

// DeleteSong removes a song; its parts and part_channels cascade away.
func (d *DB) DeleteSong(id int64) error {
	_, err := d.sql.Exec("DELETE FROM songs WHERE id=?", id)
	return err
}

// ---- parts ----

// AddPart appends a part to a song with the next sort_order and returns it with
// a full set of default channels.
func (d *DB) AddPart(songID int64, name string) (Part, error) {
	var next int
	_ = d.sql.QueryRow("SELECT COALESCE(MAX(sort_order),0)+1 FROM parts WHERE song_id=?", songID).Scan(&next)
	res, err := d.sql.Exec(
		"INSERT INTO parts(song_id, name, sort_order, master_gain) VALUES(?,?,?,?)",
		songID, name, next, 0.6)
	if err != nil {
		return Part{}, err
	}
	id, _ := res.LastInsertId()
	p := Part{ID: id, SongID: songID, Name: name, SortOrder: next, MasterGain: 0.6}
	for i := 0; i < NumChannels; i++ {
		ch := defaultChannel(i + 1)
		if err := d.SetPartChannel(id, ch); err != nil {
			return Part{}, err
		}
		p.Channels[i] = ch
	}
	return p, nil
}

func defaultChannel(n int) Channel {
	return Channel{
		Channel: n, Level: 100, Pan: 64, Enabled: n <= 2,
		KeyLow: 0, KeyHigh: 127,
		ChorusRate: 0.8, ChorusDepth: 50,
		DelayTime: 300, DelayFeedback: 30, DelayMix: 25,
	}
}

// SetPartChannel upserts one channel row of a part.
func (d *DB) SetPartChannel(partID int64, c Channel) error {
	_, err := d.sql.Exec(`
		INSERT INTO part_channels
			(part_id, channel, bank, program, instrument_name, level, pan, mute, enabled,
			 source, key_low, key_high, transpose,
			 chorus_on, chorus_rate, chorus_depth, delay_on, delay_time, delay_feedback, delay_mix)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(part_id, channel) DO UPDATE SET
			bank=excluded.bank, program=excluded.program, instrument_name=excluded.instrument_name,
			level=excluded.level, pan=excluded.pan, mute=excluded.mute, enabled=excluded.enabled,
			source=excluded.source, key_low=excluded.key_low, key_high=excluded.key_high, transpose=excluded.transpose,
			chorus_on=excluded.chorus_on, chorus_rate=excluded.chorus_rate, chorus_depth=excluded.chorus_depth,
			delay_on=excluded.delay_on, delay_time=excluded.delay_time, delay_feedback=excluded.delay_feedback, delay_mix=excluded.delay_mix`,
		partID, c.Channel, c.Bank, c.Program, c.InstrumentName, c.Level, c.Pan, b2i(c.Mute), b2i(c.Enabled),
		c.Source, c.KeyLow, c.KeyHigh, c.Transpose,
		b2i(c.ChorusOn), c.ChorusRate, c.ChorusDepth, b2i(c.DelayOn), c.DelayTime, c.DelayFeedback, c.DelayMix)
	return err
}

// ListParts returns a song's parts in order, each with its channels loaded.
func (d *DB) ListParts(songID int64) ([]Part, error) {
	rows, err := d.sql.Query(
		"SELECT id, song_id, name, sort_order, master_gain FROM parts WHERE song_id=? ORDER BY sort_order", songID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var parts []Part
	for rows.Next() {
		var p Part
		if err := rows.Scan(&p.ID, &p.SongID, &p.Name, &p.SortOrder, &p.MasterGain); err != nil {
			return nil, err
		}
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range parts {
		if err := d.loadChannels(&parts[i]); err != nil {
			return nil, err
		}
	}
	return parts, nil
}

// GetPart loads a single part (with channels) by ID.
func (d *DB) GetPart(partID int64) (Part, error) {
	var p Part
	err := d.sql.QueryRow(
		"SELECT id, song_id, name, sort_order, master_gain FROM parts WHERE id=?", partID).
		Scan(&p.ID, &p.SongID, &p.Name, &p.SortOrder, &p.MasterGain)
	if err != nil {
		return Part{}, err
	}
	return p, d.loadChannels(&p)
}

func (d *DB) loadChannels(p *Part) error {
	rows, err := d.sql.Query(`
		SELECT channel, bank, program, instrument_name, level, pan, mute, enabled,
		       source, key_low, key_high, transpose,
		       chorus_on, chorus_rate, chorus_depth, delay_on, delay_time, delay_feedback, delay_mix
		FROM part_channels WHERE part_id=? ORDER BY channel`, p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var c Channel
		var mute, en, chOn, dlOn int
		if err := rows.Scan(&c.Channel, &c.Bank, &c.Program, &c.InstrumentName, &c.Level, &c.Pan, &mute, &en,
			&c.Source, &c.KeyLow, &c.KeyHigh, &c.Transpose,
			&chOn, &c.ChorusRate, &c.ChorusDepth, &dlOn, &c.DelayTime, &c.DelayFeedback, &c.DelayMix); err != nil {
			return err
		}
		c.Mute, c.Enabled, c.ChorusOn, c.DelayOn = mute == 1, en == 1, chOn == 1, dlOn == 1
		if c.Channel >= 1 && c.Channel <= NumChannels {
			p.Channels[c.Channel-1] = c
		}
	}
	return rows.Err()
}

// RenamePart / SavePartMeta updates a part's name and master gain.
func (d *DB) SavePartMeta(partID int64, name string, masterGain float64) error {
	_, err := d.sql.Exec("UPDATE parts SET name=?, master_gain=? WHERE id=?", name, masterGain, partID)
	return err
}

// DeletePart removes a part (its channels cascade).
func (d *DB) DeletePart(partID int64) error {
	_, err := d.sql.Exec("DELETE FROM parts WHERE id=?", partID)
	return err
}

// ReorderParts sets sort_order to the position of each ID in orderedIDs.
func (d *DB) ReorderParts(orderedIDs []int64) error {
	tx, err := d.sql.Begin()
	if err != nil {
		return err
	}
	for i, id := range orderedIDs {
		if _, err := tx.Exec("UPDATE parts SET sort_order=? WHERE id=?", i+1, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// CountParts returns how many parts a song has.
func (d *DB) CountParts(songID int64) (int, error) {
	var n int
	err := d.sql.QueryRow("SELECT COUNT(*) FROM parts WHERE song_id=?", songID).Scan(&n)
	return n, err
}

// ---- app settings (key/value) ----

// SetSetting upserts a settings key.
func (d *DB) SetSetting(key, value string) error {
	_, err := d.sql.Exec(
		"INSERT INTO app_settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		key, value)
	return err
}

// GetSetting returns a settings value and whether it was present.
func (d *DB) GetSetting(key string) (string, bool, error) {
	var v string
	err := d.sql.QueryRow("SELECT value FROM app_settings WHERE key=?", key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}
