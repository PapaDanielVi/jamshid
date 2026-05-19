package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewState represents the current TUI view.
type ViewState int

const (
	ViewList ViewState = iota
	ViewConfigure
	ViewModelSelector
)

// listItem is a list item for profiles.
type listItem struct {
	title       string
	description string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.description }
func (i listItem) FilterValue() string { return i.title }

// tuimodel is the Bubble Tea model for the TUI.
type tuimodel struct {
	state         ViewState
	list          list.Model
	cfg           *Config
	cwd           string
	activeProfile string
	quitting      bool
}

// newTUI creates a new TUI model.
func newTUI(cfg *Config, cwd string) tuimodel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Jamshid Profiles"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	active := GetActiveProfile(cfg, cwd)
	m := tuimodel{
		state:         ViewList,
		list:          l,
		cfg:           cfg,
		cwd:           cwd,
		activeProfile: active,
	}

	m.refreshList()
	return m
}

func (m *tuimodel) refreshList() {
	profiles := ListProfiles(m.cfg)
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		desc := ""
		if p == m.activeProfile {
			desc = "active"
		}
		items[i] = listItem{title: p, description: desc}
	}
	m.list.SetItems(items)
}

// Init implements tea.Model.
func (m tuimodel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m tuimodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ViewList:
		return m.updateList(msg)
	case ViewConfigure:
		return m.updateConfigure(msg)
	case ViewModelSelector:
		return m.updateModelSelector(msg)
	}
	return m, nil
}

func (m tuimodel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleListKeys(&m, msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m tuimodel) updateConfigure(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.state = ViewList
			return m, nil
		}
	}
	return m, nil
}

func (m tuimodel) updateModelSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// View implements tea.Model.
func (m tuimodel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case ViewList:
		return m.listView()
	case ViewConfigure:
		return m.configureView()
	default:
		return "View not implemented\n"
	}
}

func (m tuimodel) listView() string {
	header := titleStyle.Render("Jamshid - Profile Manager")
	status := statusStyle.Render("Active: " + m.activeProfile)
	help := helpStyle.Render(keys.List)
	return lipgloss.JoinVertical(lipgloss.Left, header, status, "", m.list.View(), "", help)
}

func (m tuimodel) configureView() string {
	header := titleStyle.Render("Configure: " + m.activeProfile)
	help := helpStyle.Render(keys.Configure)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", help)
}
