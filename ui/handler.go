package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
)

type direction int

const (
	left direction = iota
	right
)

func (m model) changeFocus(direction direction) model {
	if direction == left {
		m.mode--
	} else {
		m.mode++
	}

	if m.mode < 0 {
		m.mode = 0
	}
	if m.mode > container {
		m.mode = container
	}

	for i := range m.columns {
		m.columns[i].isFocused = i == int(m.mode)
	}

	return m
}

func handleRightKey(m model) (tea.Model, tea.Cmd) {
	return handleEnterKey(m)
}

func handleLeftKey(m model) (tea.Model, tea.Cmd) {
	return m.changeFocus(left), nil
}

func handleEnterKey(m model) (model, tea.Cmd) {
	switch {
	case m.columns[namespace].isFocused:
		return m.onAction(namespaceSelect)

	case m.columns[pod].isFocused:
		return m.onAction(podSelect)

	default:
		return m.onAction(containerSelect)
	}
}

func handlek8sMsg(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case namespaceMsg:
		items := createDisplayedListFromMetadata(msg.val.list.Items, func(nms v1.Namespace) DiplayableItemList {
			return &displayableMeta{&nms.ObjectMeta}
		})
		m.columns[namespace], cmd = m.columns[namespace].Update(
			listUpdateMsg{
				statusTxt:        "namespace",
				title:            "namespace",
				val:              items,
				preSelectedValue: msg.val.preSelectedNamespace,
			},
		)

	case podMsg:
		items := createDisplayedListFromMetadata(msg.val.list.Items, func(nms v1.Pod) DiplayableItemList {
			return &displayableMeta{&nms.ObjectMeta}
		})
		m.columns[pod], cmd = m.columns[pod].Update(
			listUpdateMsg{
				statusTxt:        "pod",
				title:            "Pod",
				val:              items,
				preSelectedValue: msg.val.preSelectedPod,
			},
		)

	case containerMsg:
		items := createDisplayedListFromMetadata(msg.val.list, func(container v1.Container) DiplayableItemList {
			return &displayableContainer{container}
		})
		m.columns[container], cmd = m.columns[container].Update(
			listUpdateMsg{
				statusTxt:        "hehe",
				title:            "Container",
				val:              items,
				preSelectedValue: msg.val.preSelectedContainer,
			},
		)
	}

	return m, cmd
}
