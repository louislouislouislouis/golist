package ui

import (
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/louislouislouislouis/repr8ducer/utils"
)

type Action int

const (
	podSelect Action = iota
	containerSelect
	namespaceSelect
)

func (m model) onAction(a Action) (model, tea.Cmd) {
	var cmd tea.Cmd

	if len(m.columns[m.mode].list.VisibleItems()) == 0 {
		return m, nil
	}
	switch a {
	case podSelect:
		m.columns[pod].current = selectTitleSelected(m.columns[pod].list)
		return m, initContainers(m.columns[namespace].current, m.columns[pod].current)
	case containerSelect:
		m.columns[container].current = selectTitleSelected(m.columns[container].list)
		// Todo : Handle error
		command, err := m.k8sService.Exec(
			m.columns[namespace].current,
			m.columns[pod].current,
			m.columns[container].current,
		)
		if err != nil {
			utils.Log.Error().Msg(err.Error())
		}

		clipboard.WriteAll(command)
	case namespaceSelect:
		m.columns[namespace].current = selectTitleSelected(m.columns[namespace].list)
		return m, initPods(m.columns[namespace].current)

	}
	return m, cmd
}

func selectTitleSelected(list list.Model) string {
	return list.VisibleItems()[list.Index()].(displayedItem).title
}
