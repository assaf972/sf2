package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// S07-T02: reordering parts persists their sort_order.
func TestReorderPartsPersists(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "ro.db"))
	require.NoError(t, err)
	defer d.Close()

	s, _ := d.CreateSong("S")
	a, _ := d.AddPart(s.ID, "A")
	b, _ := d.AddPart(s.ID, "B")
	c, _ := d.AddPart(s.ID, "C")

	// Reorder to C, A, B.
	require.NoError(t, d.ReorderParts([]int64{c.ID, a.ID, b.ID}))

	parts, err := d.ListParts(s.ID)
	require.NoError(t, err)
	require.Len(t, parts, 3)

	var names []string
	for i, p := range parts {
		names = append(names, p.Name)
		require.Equal(t, i+1, p.SortOrder, "sort_order should be contiguous from 1")
	}
	require.Equal(t, []string{"C", "A", "B"}, names)
}
