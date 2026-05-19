package tui

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/PapaDanielVi/jamshid/internal/pkg/config"
	"github.com/PapaDanielVi/jamshid/internal/pkg/profile"
)

type ViewState int

const (
	ViewCommands ViewState = iota
	ViewProfiles
	ViewTextInput
	ViewVaultSubcommands
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
	textInput       textinput.Model
	cfg             *config.Config
	cwd             string
	activeProfile   string
	quitting        bool
	selectedCmd     string
	selectedProfile string
	selectedSubcmd  string
	textInputPrompt string
	focusTextInput  bool
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

var vaultSubcommands = []listItem{
	{title: "init", description: "Initialize git vault with remote URL"},
	{title: "sync", description: "Sync vault with remote"},
}

func NewTUI(cfg *config.Config, cwd string) tuiModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Jamshid Commands"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 128
	ti.SetWidth(40)

	active := profile.GetActiveProfile(cfg, cwd)
	m := tuiModel{
		state:         ViewCommands,
		list:          l,
		textInput:     ti,
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

func (m *tuiModel) setVaultSubcommandItems() {
	items := make([]list.Item, len(vaultSubcommands))
	for i, sc := range vaultSubcommands {
		items[i] = sc
	}
	_ = m.list.SetItems(items)
	m.list.Title = "Vault Subcommand"
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

	if m.focusTextInput {
		m.focusTextInput = false
		return m, m.textInput.Focus()
	}

	switch m.state {
	case ViewCommands:
		return m.updateCommands(msg)
	case ViewProfiles:
		return m.updateProfiles(msg)
	case ViewTextInput:
		return m.updateTextInput(msg)
	case ViewVaultSubcommands:
		return m.updateVaultSubcommands(msg)
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
						m.state = ViewTextInput
						m.textInputPrompt = "Profile name"
						m.textInput.Placeholder = "Enter profile name..."
						m.focusTextInput = true
						return m, nil
					case "delete", "link", "env":
						if len(profile.ListProfiles(m.cfg)) == 0 {
							m.quitting = true
							return m, tea.Quit
						}
						m.state = ViewProfiles
						m.setProfileItems()
						return m, nil
					case "list", "unlink", "help":
						return m, tea.Quit
					case "vault":
						m.state = ViewVaultSubcommands
						m.setVaultSubcommandItems()
						return m, nil
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

func (m *tuiModel) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			val := m.textInput.Value()
			if val != "" {
				m.selectedProfile = val
				return m, tea.Quit
			}
		}
	}

	m.textInput, _ = m.textInput.Update(msg)
	return m, nil
}

func (m *tuiModel) updateVaultSubcommands(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.selectedSubcmd = item.title
					if item.title == "init" {
						m.state = ViewTextInput
						m.textInputPrompt = "Remote URL"
						m.textInput.Placeholder = "Enter git remote URL..."
						m.focusTextInput = true
						return m, nil
					}
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
		return tea.NewView("")
	}

	var content string
	switch m.state {
	case ViewCommands:
		content = m.commandsView()
	case ViewProfiles:
		content = m.profilesView()
	case ViewTextInput:
		content = m.textInputView()
	case ViewVaultSubcommands:
		content = m.vaultSubcommandsView()
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

func (m tuiModel) textInputView() string {
	header := titleStyle.Render("Jamshid - " + m.textInputPrompt)
	help := helpStyle.Render("enter: confirm, esc: back, q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.textInput.View(), "", help)
}

func (m tuiModel) vaultSubcommandsView() string {
	header := titleStyle.Render("Jamshid - Vault")
	help := helpStyle.Render("↑/↓: navigate, enter: select, esc: back, q: quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", m.list.View(), "", help)
}

// SelectedCommand returns the selected command, its argument (profile name or
// text input value), and an optional subcommand.
func (m tuiModel) SelectedCommand() (cmd string, arg string, subcmd string) {
	return m.selectedCmd, m.selectedProfile, m.selectedSubcmd
}
