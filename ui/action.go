package ui

import (
	"context"

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
		utils.Log.Debug().Msgf("%d", m.mode)
		m.columns[pod].current = selectTitleSelected(m.columns[pod].list)
		cmd = initContainers(m.columns[namespace].current, m.columns[pod].current, "", context.TODO())
	case containerSelect:
		m.columns[container].current = selectTitleSelected(m.columns[container].list)
		command, err := m.k8sService.PodToContainer(
			m.columns[namespace].current,
			m.columns[pod].current,
			context.TODO(),
		)
		if err != nil {
			utils.Log.Error().Msg(err.Error())
			cmd = updateStatusLine(err.Error())
		}
		clipboard.WriteAll(command)
	case namespaceSelect:
		m.columns[namespace].current = selectTitleSelected(m.columns[namespace].list)
		cmd = initPods(m.columns[namespace].current, "", context.TODO())

	}
	m = m.changeFocus(right)
	return m, cmd
}

func selectTitleSelected(list list.Model) string {
	return list.VisibleItems()[list.Index()].(displayedItem).title
}
