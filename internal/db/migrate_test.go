package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// S02-T02: opening a pre-FX database adds the chorus_*/delay_* columns without
// losing existing data.
func TestMigrateAddsFXColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")

	// 1) Hand-build an "old" database whose part_channels has NO effects columns.
	raw, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	_, err = raw.Exec(`
		CREATE TABLE songs(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);
		CREATE TABLE parts(id INTEGER PRIMARY KEY AUTOINCREMENT, song_id INTEGER, name TEXT, sort_order INTEGER, master_gain REAL);
		CREATE TABLE part_channels(part_id INTEGER, channel INTEGER, bank INTEGER DEFAULT 0, program INTEGER DEFAULT 0,
			instrument_name TEXT DEFAULT '', level INTEGER DEFAULT 100, pan INTEGER DEFAULT 64, mute INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 1, source TEXT DEFAULT '', key_low INTEGER DEFAULT 0, key_high INTEGER DEFAULT 127,
			transpose INTEGER DEFAULT 0, PRIMARY KEY(part_id, channel));
		CREATE TABLE midi_keyboards(id INTEGER PRIMARY KEY AUTOINCREMENT, manufacturer TEXT, model TEXT, midi_mapping TEXT);
		INSERT INTO songs(id,name) VALUES(1,'Legacy');
		INSERT INTO parts(id,song_id,name,sort_order,master_gain) VALUES(1,1,'OldPart',1,0.6);
		INSERT INTO part_channels(part_id,channel,level) VALUES(1,1,111);`)
	require.NoError(t, err)
	require.NoError(t, raw.Close())

	// 2) Open via our migrating Open() — should add FX columns, keep the old row.
	d, err := Open(path)
	require.NoError(t, err)
	defer d.Close()

	for _, col := range []string{"chorus_on", "chorus_rate", "chorus_depth", "delay_on", "delay_time", "delay_feedback", "delay_mix"} {
		require.NoError(t, d.addColumnIfMissing("part_channels", col, "INTEGER"),
			"column %s should already exist after migrate", col)
	}

	p, err := d.GetPart(1)
	require.NoError(t, err)
	require.Equal(t, "OldPart", p.Name)
	require.Equal(t, 111, p.Channels[0].Level, "old data preserved")
	require.Equal(t, 0.8, p.Channels[0].ChorusRate, "FX default applied")
	require.False(t, p.Channels[0].ChorusOn)
}
