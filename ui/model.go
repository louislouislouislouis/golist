package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/louislouislouislouis/repr8ducer/k8s"
)

type itemWithList struct {
	current       string
	list          list.Model
	isInitialized bool
}

type size struct {
	width  int
	height int
}
type model struct {
	columns    []column
	mode       colType
	k8sService *k8s.K8sService
	spinner    spinner.Model
	size       size
}

type ModelConfig struct {
	Namespace, Pod, Container string
}

type colType int

const (
	namespace colType = iota
	pod
	container
)

func NewModel(k8s *k8s.K8sService, c ModelConfig) model {
	spinners := spinner.New()
	test := []column{
		{
			width:         0,
			height:        0,
			current:       c.Namespace,
			isInitialized: false,
			isFocused:     true,
			spinner:       spinners,
			list:          setupCustomList("Namespace", []list.Item{}),
		},
		{
			width:         0,
			isInitialized: false,
			current:       c.Pod,
			height:        0,
			isFocused:     false,
			spinner:       spinners,
			list:          setupCustomList("Pods", []list.Item{}),
		},
		{
			width:         0,
			isInitialized: false,
			height:        0,
			spinner:       spinners,
			current:       c.Container,
			isFocused:     false,
			list:          setupCustomList("Containers", []list.Item{}),
		},
	}
	return model{
		columns:    test,
		k8sService: k8s,
	}
}

func setupCustomList(title string, items []list.Item) list.Model {
	theList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	theList.StatusMessageLifetime = time.Hour
	theList.SetShowHelp(false)
	theList.Title = title
	return theList
}

func (m model) Init() tea.Cmd {
	var initCmd tea.Cmd
	if m.columns[namespace].current == "" {
		initCmd = initNamespace
	} else if m.columns[pod].current == "" {
		initCmd = initPods(m.columns[namespace].current)
	} else if m.columns[container].current == "" {
		initCmd = initContainers(m.columns[namespace].current, m.columns[pod].current)
	}

	return tea.Batch(
		m.columns[pod].spinner.Tick,
		m.columns[container].spinner.Tick,
		m.columns[namespace].spinner.Tick,
		initCmd,
	)
}

func title() string {
	title := ` ____  ___  ___  ____   
(_  _)(  _)/ __)(_  _)  
  )(   ) _)\__ \  )(    
 (__) (___)(___/ (__)`
	return title
}

func (m model) View() string {
	renders := make([]string, len(m.columns))
	for _, c := range m.columns {
		renders = append(renders, c.View())
	}
	m.columns[m.mode].list.Help.Width = 444
	return lipgloss.JoinVertical(
		lipgloss.Center,
		bigTitleStyle.Width(m.size.width).Render(title()),
		lipgloss.JoinHorizontal(lipgloss.Left, renders...),
		m.columns[m.mode].list.Help.View(m.columns[m.mode].list),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.columns))
	switch msg := msg.(type) {

	case namespaceMsg, podMsg, containerMsg:
		return handlek8sMsg(m, msg)

	case tea.KeyMsg:
		if m.isFiltering() {
			break
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			return handleEnterKey(m)

		case "left":
			return handleLeftKey(m)

		case "right":
			return handleRightKey(m)

		}

	case tea.WindowSizeMsg:
		m.size = size{
			height: msg.Height,
			width:  msg.Width,
		}
		for i := range m.columns {
			col, cmd := m.columns[i].Update(msg)
			m.columns[i] = col
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		for i := range m.columns {
			col, cmd := m.columns[i].Update(msg)
			m.columns[i] = col
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	m.columns[m.mode], cmd = m.columns[m.mode].Update(msg)
	return m, tea.Batch(cmd)
}

func (m model) isFiltering() bool {
	for i := range m.columns {
		if m.columns[i].list.FilterState() == list.Filtering {
			return true
		}
	}
	return false
}
