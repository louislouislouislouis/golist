package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/louislouislouislouis/repr8ducer/k8s"
)

func initNamespace(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		listNamespace, _ := k8s.GetService().ListNamespace(ctx)
		return namespaceMsg{
			val: listNamespace,
		}
	}
}

func initContainers(nms, pod string, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		containers, _ := k8s.GetService().GetContainerFromPods(nms, pod, ctx)
		return containerMsg{
			val: containers,
		}
	}
}

func initPods(nms string, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		listPods, _ := k8s.GetService().ListPodsInNamespace(nms, ctx)
		return podMsg{
			val: listPods,
		}
	}
}
