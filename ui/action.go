package ui

import (
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type Action int

const (
	podSelect Action = iota
	containerSelect
	namespaceSelect
)

func (m model) onAction(a Action) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a {
	case podSelect:
		m.pod = selectTitleSelected(m.listPods)
		return m, initContainers(m.namespace, m.pod)
		//		containers, _ := m.k8sService.GetContainerFromPods(m.pod, m.namespace)
		//		displayedContainerList := createDisplayedListFromMetadata(
		//			containers,
		//			func(c v1.Container) DiplayableItemList {
		//				return &displayableContainer{c}
		//			},
		//		)
		//
		//		cmd = tea.Batch(
		//			m.listContainer.SetItems(displayedContainerList),
		//			m.listContainer.NewStatusMessage(
		//				fmt.Sprint(
		//					statusStyle.Render("You're seeing containers in pod: "),
		//					statusFocusStyle.Render(m.pod),
		//					statusStyle.Render(". Select a container ⬇ "),
		//				),
		//			),
		//		)

	case containerSelect:
		m.container = selectTitleSelected(m.listContainer)
		// Todo : Handle error
		command, _ := m.k8sService.Exec(m.namespace, m.pod, m.container)
		clipboard.WriteAll(command)
	case namespaceSelect:
		m.namespace = selectTitleSelected(m.listNamespace)
		return m, initPods(m.namespace)

		//		m.namespace = selectTitleSelected(m.listNamespace)
		//		// TODO : Handle error
		//		pods, _ := m.k8sService.ListPodsInNamespace(m.namespace)
		//		cmd = tea.Batch(
		//			m.listPods.SetItems(
		//				createDisplayedListFromMetadata(
		//					pods.Items,
		//					func(pod v1.Pod) DiplayableItemList {
		//						return &displayableMeta{&pod}
		//					}),
		//			),
		//			m.listPods.NewStatusMessage(
		//				fmt.Sprint(
		//					statusStyle.Render("You're seeing pods in namespace: "),
		//					statusFocusStyle.Render(m.namespace),
		//					statusStyle.Render(". Select a pod ⬇ "),
		//				),
		//			),
		//		)
	}
	return m, cmd
}

func selectTitleSelected(list list.Model) string {
	// TODO handle if index not in range
	return list.Items()[list.Index()].(displayedItem).title
}
