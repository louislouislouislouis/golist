package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/louislouislouislouis/repr8ducer/k8s"
)

func initNamespace() tea.Msg {
	listNamespace, _ := k8s.GetService().ListNamespace()
	return namespaceMsg{
		val: listNamespace,
	}
}

func initContainers(nms, pod string) tea.Cmd {
	return func() tea.Msg {
		containers, _ := k8s.GetService().GetContainerFromPods(nms, pod)
		return containerMsg{
			val: containers,
		}
	}
}

func initPods(nms string) tea.Cmd {
	return func() tea.Msg {
		listPods, _ := k8s.GetService().ListPodsInNamespace(nms)
		return podMsg{
			val: listPods,
		}
	}
}
