package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/gitignore"
	"github.com/PapaDanielVi/jamshid/internal/pkg/profile"
)

// keyMap defines the key bindings for the TUI.
type keyMap struct {
	List      string
	Configure string
}

var keys = keyMap{
	List:      "↑/↓: navigate, enter: select global, l: link, u: unlink, c: configure, q: quit",
	Configure: "tab: next field, enter: save, esc: back",
}

// handleListKeys handles key presses in list view.
func handleListKeys(m *tuimodel, msg tea.KeyMsg) (tuimodel, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return *m, tea.Quit
	case "enter":
		selected := m.list.SelectedItem()
		if selected != nil {
			if p, ok := selected.(listItem); ok {
				m.activeProfile = p.title
				m.cfg.GlobalProfile = p.title
				_ = config.SaveConfig(m.cfg)
				m.refreshList()
			}
		}
	case "l":
		selected := m.list.SelectedItem()
		if selected != nil {
			if p, ok := selected.(listItem); ok {
				if profile.IsGitRepo(m.cwd) {
					_ = profile.LinkProfile(m.cfg, m.cwd, p.title, false)
					_ = gitignore.EnsureGitignore(m.cwd)
					_ = config.SaveConfig(m.cfg)
					m.activeProfile = p.title
					m.refreshList()
				}
			}
		}
	case "u":
		_ = profile.UnlinkProfile(m.cfg, m.cwd)
		_ = config.SaveConfig(m.cfg)
		m.activeProfile = m.cfg.GlobalProfile
		m.refreshList()
	case "c":
		// Enter configure mode for active profile
		if m.activeProfile != "" {
			_, ok := profile.GetProfile(m.cfg, m.activeProfile)
			if ok {
				m.state = ViewConfigure
			}
		}
	}
	return *m, nil
}
