package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"

	"github.com/louislouislouislouis/repr8ducer/k8s"
)

type model struct {
	k8sService    *k8s.K8sService
	namespace     string
	pod           string
	container     string
	listNamespace list.Model
	listContainer list.Model
	listPods      list.Model
}

type ModelConfig struct {
	Namespace, Pod, Container string
}

func NewModel(k8s *k8s.K8sService, c ModelConfig) model {
	displayedNamespaceList := []list.Item{}
	displayedPodList := []list.Item{}
	displayedContainerList := []list.Item{}

	return model{
		k8sService:    k8s,
		pod:           c.Pod,
		namespace:     c.Namespace,
		container:     c.Container,
		listNamespace: setupCustomList("Namespaces", displayedNamespaceList),
		listPods:      setupCustomList("Pods", displayedPodList),
		listContainer: setupCustomList("Containers", displayedContainerList),
	}
}

func setupCustomList(title string, items []list.Item) list.Model {
	theList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	theList.StatusMessageLifetime = time.Hour
	theList.Title = title
	return theList
}

type (
	errMsg int
)

type namespaceMsg struct {
	val *v1.NamespaceList
}

type podMsg struct {
	val *v1.PodList
}

type containerMsg struct {
	val []v1.Container
}

func (m model) Init() tea.Cmd {
	if m.namespace == "" {
		return initNamespace
	}
	if m.pod == "" {
		return initPods(m.namespace)
	}

	if m.container == "" {
		return initContainers(m.namespace, m.pod)
	}
	return nil
}

func initNamespace() tea.Msg {
	listNamespace, _ := k8s.GetService().ListNamespace()
	return namespaceMsg{
		val: listNamespace,
	}
}

func initContainers(nms, pod string) tea.Cmd {
	return func() tea.Msg {
		container, _ := k8s.GetService().GetContainerFromPods(nms, pod)
		return containerMsg{
			val: container,
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

func (m model) View() string {
	if m.namespace == "" {
		return docStyle.Render(m.listNamespace.View())
	}

	if m.pod == "" {
		return docStyle.Render(m.listPods.View())
	}

	if m.container == "" {
		return docStyle.Render(m.listContainer.View())
	}

	return fmt.Sprint(
		"Selected ",
		orangeStyle.Render(m.namespace),
		"/",
		blueStyle.Render(m.pod),
		":",
		violetStyle.Render(m.container),
	)
}

func (m model) isFiltering() bool {
	return m.listNamespace.FilterState() == list.Filtering ||
		m.listContainer.FilterState() == list.Filtering ||
		m.listPods.FilterState() == list.Filtering
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.isFiltering() {
			break
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			if m.namespace == "" {
				return m.onAction(namespaceSelect)
			}
			if m.pod == "" {
				return m.onAction(podSelect)
			}
			return m.onAction(containerSelect)
		}
	case tea.WindowSizeMsg:
		m = updateSizeLists(m, msg)
	case namespaceMsg:
		displayedNamespaceList := createDisplayedListFromMetadata(
			msg.val.Items,
			func(nms v1.Namespace) DiplayableItemList {
				return &displayableMeta{&nms.ObjectMeta}
			},
		)
		return m, m.listNamespace.SetItems(displayedNamespaceList)
	case podMsg:
		displayedPodList := createDisplayedListFromMetadata(
			msg.val.Items,
			func(pod v1.Pod) DiplayableItemList {
				return &displayableMeta{&pod.ObjectMeta}
			},
		)
		return m, m.listPods.SetItems(displayedPodList)
	case containerMsg:
		displayedContainerList := createDisplayedListFromMetadata(
			msg.val,
			func(container v1.Container) DiplayableItemList {
				return &displayableContainer{container}
			},
		)
		return m, m.listContainer.SetItems(displayedContainerList)
	}
	return updateList(m, msg)
}

func updateList(m model, msg tea.Msg) (model, tea.Cmd) {
	var cmdNamespace, cmdPods, cmdContainer tea.Cmd
	m.listNamespace, cmdNamespace = m.listNamespace.Update(msg)
	m.listPods, cmdPods = m.listPods.Update(msg)
	m.listContainer, cmdContainer = m.listContainer.Update(msg)
	return m, tea.Batch(cmdPods, cmdNamespace, cmdContainer)
}

func updateSizeLists(m model, msg tea.WindowSizeMsg) model {
	h, v := docStyle.GetFrameSize()
	m.listNamespace.SetSize(msg.Width-h, msg.Height-v)
	m.listContainer.SetSize(msg.Width-h, msg.Height-v)
	m.listPods.SetSize(msg.Width-h, msg.Height-v)
	return m
}
