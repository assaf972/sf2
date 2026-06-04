package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// S03-T01: the shell builds an AppTabs with the three named views in order.
func TestShellHasThreeTabs(t *testing.T) {
	u, _ := newTestUI(t)
	tabs := u.buildShell()

	var labels []string
	for _, it := range tabs.Items {
		labels = append(labels, it.Text)
	}
	assert.Equal(t, []string{"Live", "Songs", "Settings"}, labels)
}
