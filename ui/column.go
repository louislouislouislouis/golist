package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type column struct {
	width         int
	height        int
	isFocused     bool
	isInitialized bool
	list          list.Model
	current       string
	spinner       spinner.Model
}

func (c column) Init() tea.Cmd {
	return nil
}

func (c column) View() string {
	var text string
	if c.isInitialized {
		text = c.list.View()
	} else {
		text = c.spinner.View()
	}

	if c.isFocused {
		return focusStyle.Render(text)
	}
	return unfocusStyle.Render(text)
}

func (c column) Update(msg tea.Msg) (column, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		c.list.SetSize((msg.Width-3*8)/3, msg.Height/2)

	case listUpdateMsg:
		c.isInitialized = true
		c.isFocused = true
		cmds := tea.Batch(
			c.list.SetItems(msg.val),
			c.list.NewStatusMessage(msg.statusTxt),
		)
		return c, cmds

	case spinner.TickMsg:
		var cmd tea.Cmd
		c.spinner, cmd = c.spinner.Update(msg)
		return c, cmd

	}

	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}
