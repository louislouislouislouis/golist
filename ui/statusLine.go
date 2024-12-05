package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type statusLine struct {
	text string
}

func (m statusLine) Init() tea.Cmd {
	return nil
}

func (m statusLine) Update(msg tea.Msg) (statusLine, tea.Cmd) {
	switch msg := msg.(type) {
	case infoMsg:
		m.text = msg.val
	}
	return m, nil
}

func (m statusLine) View() string {
	return m.text
}

func updateStatusLine(msg string) tea.Cmd {
	return func() tea.Msg {
		return infoMsg{
			val: msg,
		}
	}
}
