package tui

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/profile"
)

type ViewState int

const (
	ViewCommands ViewState = iota
	ViewProfiles
)

type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

type tuiModel struct {
	state           ViewState
	list            list.Model
	cfg             *config.Config
	cwd             string
	activeProfile   string
	quitting        bool
	selectedCmd     string
	selectedProfile string
	width           int
	height          int
}

var commands = []listItem{
	{title: "add", description: "Create a new profile"},
	{title: "delete", description: "Delete a profile"},
	{title: "list", description: "List all profiles"},
	{title: "link", description: "Link profile to current directory"},
	{title: "unlink", description: "Unlink profile from current directory"},
	{title: "env", description: "Set CLAUDE_CONFIG_DIR for a profile"},
	{title: "vault", description: "Manage Git vault"},
	{title: "help", description: "Show help"},
}

func NewTUI(cfg *config.Config, cwd string) tuiModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Jamshid Commands"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	active := profile.GetActiveProfile(cfg, cwd)
	m := tuiModel{
		state:         ViewCommands,
		list:          l,
		cfg:           cfg,
		cwd:           cwd,
		activeProfile: active,
	}

	m.setCommandItems()
	return m
}

func (m *tuiModel) setCommandItems() {
	items := make([]list.Item, len(commands))
	for i, cmd := range commands {
		items[i] = cmd
	}
	_ = m.list.SetItems(items)
	m.list.Title = "Jamshid Commands"
}

func (m *tuiModel) setProfileItems() {
	profiles := profile.ListProfiles(m.cfg)
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		desc := ""
		if p == m.activeProfile {
			desc = "active"
		}
		items[i] = listItem{title: p, description: desc}
	}
	_ = m.list.SetItems(items)
	m.list.Title = "Select Profile"
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	}

	switch m.state {
	case ViewCommands:
		return m.updateCommands(msg)
	case ViewProfiles:
		return m.updateProfiles(msg)
	default:
		return m, nil
	}
}

func (m *tuiModel) updateCommands(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				if item, ok := selected.(listItem); ok {
					m.selectedCmd = item.title
					switch item.title {
					case "add":
						return m, tea.Quit
					case "delete", "link", "env", "list":
						m.state = ViewProfiles
						m.setProfileItems()
						return m, nil
					default:
						return m, tea.Quit
					}
				}
			}
		}
	}

	m.list, _ = m.list.Update(msg)
	return m, nil
}

func (m *tuiModel) updateProfiles(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.state = ViewCommands
			m.setCommandItems()
			return m, nil
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				if item, ok := selected.(listItem); ok {
					m.selectedProfile = item.title
					return m, tea.Quit
				}
			}
		}
	}

	m.list, _ = m.list.Update(msg)
	return m, nil
}

func (m tuiModel) View() tea.View {
	if m.quitting {
		v := tea.NewView("Goodbye!\n")
		return v
	}

	var content string
	switch m.state {
	case ViewCommands:
		content = m.commandsView()
	case ViewProfiles:
		content = m.profilesView()
	default:
		content = "View not implemented\n"
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func (m tuiModel) commandsView() string {
	header := titleStyle.Render("Jamshid - Command Selection")
	help := helpStyle.Render("↑/↓: navigate, enter: select, q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.list.View(), "", help)
}

func (m tuiModel) profilesView() string {
	header := titleStyle.Render("Jamshid - Profile Selection")
	help := helpStyle.Render("↑/↓: navigate, enter: select, esc: back, q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.list.View(), "", help)
}

func (m tuiModel) SelectedCommand() (string, string) {
	return m.selectedCmd, m.selectedProfile
}
