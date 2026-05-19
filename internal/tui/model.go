package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/profile"
)

// ViewState represents the current TUI view.
type ViewState int

const (
	ViewCommands ViewState = iota
	ViewProfiles
	ViewConfigure
	ViewModelSelector
)

// listItem is a list item for profiles or commands.
type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

// tuimodel is the Bubble Tea model for the TUI.
type tuimodel struct {
	state          ViewState
	list           list.Model
	cfg            *config.Config
	cwd            string
	activeProfile  string
	quitting       bool
	selectedCmd    string
	selectedProfile string
}

// Command definitions for TUI.
var commands = []listItem{
	{title: "add", description: "Create a new profile"},
	{title: "delete", description: "Delete a profile"},
	{title: "list", description: "List all profiles"},
	{title: "link", description: "Link profile to current directory"},
	{title: "unlink", description: "Unlink profile from current directory"},
	{title: "global", description: "Set global profile"},
	{title: "vault", description: "Manage Git vault"},
	{title: "help", description: "Show help"},
}

// NewTUI creates a new TUI model.
func NewTUI(cfg *config.Config, cwd string) tuimodel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Jamshid Commands"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	active := profile.GetActiveProfile(cfg, cwd)
	m := tuimodel{
		state:         ViewCommands,
		list:          l,
		cfg:           cfg,
		cwd:           cwd,
		activeProfile: active,
	}

	m.setCommandItems()
	return m
}

func (m *tuimodel) setCommandItems() {
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = cmd
	}
	m.list.SetItems(items)
	m.list.Title = "Jamshid Commands"
}

func (m *tuimodel) setProfileItems() {
	profiles := profile.ListProfiles(m.cfg)
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		desc := ""
		if p == m.activeProfile {
			desc = "active"
		}
		items[i] = listItem{title: p, description: desc}
	}
	m.list.SetItems(items)
	m.list.Title = "Select Profile"
}

func (m *tuimodel) refreshList() {
	if m.state == ViewCommands {
		m.setCommandItems()
	} else if m.state == ViewProfiles {
		m.setProfileItems()
	}
}

// Init implements tea.Model.
func (m tuimodel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m tuimodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ViewCommands:
		return m.updateCommands(msg)
	case ViewProfiles:
		return m.updateProfiles(msg)
	case ViewConfigure:
		return m.updateConfigure(msg)
	case ViewModelSelector:
		return m.updateModelSelector(msg)
	}
	return m, nil
}

func (m *tuimodel) updateCommands(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return *m, tea.Quit
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				if item, ok := selected.(listItem); ok {
					m.selectedCmd = item.title
					// If command needs profile selection, switch to profile view
					switch item.title {
					case "add", "delete", "link", "global":
						if item.title == "add" {
							// Add command doesn't need profile selection, prompt for name
							return *m, tea.Quit
						}
						m.state = ViewProfiles
						m.setProfileItems()
						return *m, nil
					default:
						// Other commands exit TUI and execute
						return *m, tea.Quit
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *tuimodel) updateProfiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return *m, tea.Quit
		case "esc":
			m.state = ViewCommands
			m.setCommandItems()
			return *m, nil
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				if item, ok := selected.(listItem); ok {
					m.selectedProfile = item.title
					return *m, tea.Quit
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *tuimodel) updateConfigure(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.state = ViewProfiles
			return m, nil
		}
	}
	return m, nil
}

func (m *tuimodel) updateModelSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model.
func (m tuimodel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case ViewCommands:
		return m.commandsView()
	case ViewProfiles:
		return m.profilesView()
	case ViewConfigure:
		return m.configureView()
	default:
		return "View not implemented\n"
	}
}

func (m tuimodel) commandsView() string {
	header := titleStyle.Render("Jamshid - Command Selection")
	help := helpStyle.Render(keys.List)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.list.View(), "", help)
}

func (m tuimodel) profilesView() string {
	header := titleStyle.Render("Jamshid - Profile Selection")
	help := helpStyle.Render("↑/↓: navigate, enter: select, esc: back, q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.list.View(), "", help)
}

func (m tuimodel) configureView() string {
	header := titleStyle.Render("Configure: " + m.activeProfile)
	help := helpStyle.Render(keys.Configure)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", help)
}

// SelectedCommand returns the selected command and profile.
func (m tuimodel) SelectedCommand() (string, string) {
	return m.selectedCmd, m.selectedProfile
}
