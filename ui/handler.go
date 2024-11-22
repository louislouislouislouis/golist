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

func changeFocus(m model, direction direction) (model, tea.Cmd) {
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

	return m, nil
}

func handleRightKey(m model) (tea.Model, tea.Cmd) {
	return handleEnterKey(m)
}

func handleLeftKey(m model) (tea.Model, tea.Cmd) {
	return changeFocus(m, left)
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
		items := createDisplayedListFromMetadata(msg.val.Items, func(nms v1.Namespace) DiplayableItemList {
			return &displayableMeta{&nms.ObjectMeta}
		})
		m.mode = namespace
		m.columns[namespace], cmd = m.columns[namespace].Update(
			listUpdateMsg{
				statusTxt: "namespace",
				title:     "namespace",
				val:       items,
			},
		)

	case podMsg:
		items := createDisplayedListFromMetadata(msg.val.Items, func(nms v1.Pod) DiplayableItemList {
			return &displayableMeta{&nms.ObjectMeta}
		})
		m.columns[pod], cmd = m.columns[pod].Update(
			listUpdateMsg{
				statusTxt: "pod",
				title:     "Pod",
				val:       items,
			},
		)
		m.mode = pod
		m.columns[namespace].isFocused = false
		m.columns[container].isFocused = false

	case containerMsg:
		items := createDisplayedListFromMetadata(msg.val, func(container v1.Container) DiplayableItemList {
			return &displayableContainer{container}
		})
		m.columns[container], cmd = m.columns[container].Update(
			listUpdateMsg{
				statusTxt: "hehe",
				title:     "Container",
				val:       items,
			},
		)
		m.mode = container
		m.columns[namespace].isFocused = false
		m.columns[pod].isFocused = false
	}

	return m, cmd
}
