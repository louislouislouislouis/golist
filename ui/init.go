package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/louislouislouislouis/repr8ducer/k8s"
)

func initNamespace(preSelectedNamepace string, ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		// Todo hadle error
		listNamespace, _ := k8s.GetService().ListNamespace(ctx)
		return namespaceMsg{
			val: namespaceMsgValue{
				list:                 listNamespace,
				preSelectedNamespace: preSelectedNamepace,
			},
		}
	}
}

func initContainers(nms, pod, preSelectedContainer string, ctx context.Context) tea.Cmd {
	// Todo hadle error
	return func() tea.Msg {
		containers, _ := k8s.GetService().GetContainerFromPods(nms, pod, ctx)
		return containerMsg{
			val: containerMsgValue{
				list:                 containers,
				preSelectedContainer: preSelectedContainer,
			},
		}
	}
}

func initPods(nms, preSelectedPod string, ctx context.Context) tea.Cmd {
	// Todo hadle error
	return func() tea.Msg {
		listPods, _ := k8s.GetService().ListPodsInNamespace(nms, ctx)
		return podMsg{
			val: podMsgValue{
				list:           listPods,
				preSelectedPod: preSelectedPod,
			},
		}
	}
}
