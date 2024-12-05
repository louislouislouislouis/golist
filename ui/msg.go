package ui

import (
	"github.com/charmbracelet/bubbles/list"
	v1 "k8s.io/api/core/v1"
)

type statusMsg int

const (
	error statusMsg = iota
	ok
)

type namespaceMsg struct {
	val    *v1.NamespaceList
	status statusMsg
}

type infoMsg struct {
	val    string
	status statusMsg
}

type podMsg struct {
	val    *v1.PodList
	status statusMsg
}

type containerMsg struct {
	val    []v1.Container
	status statusMsg
}

type listUpdateMsg struct {
	val       []list.Item
	title     string
	statusTxt string
	status    statusMsg
}
