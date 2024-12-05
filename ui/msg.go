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

type namespaceMsgValue struct {
	list                 *v1.NamespaceList
	preSelectedNamespace string
}

type containerMsgValue struct {
	list                 []v1.Container
	preSelectedContainer string
}
type podMsgValue struct {
	list           *v1.PodList
	preSelectedPod string
}

type namespaceMsg struct {
	val    namespaceMsgValue
	status statusMsg
}

type infoMsg struct {
	val    string
	status statusMsg
}

type podMsg struct {
	val    podMsgValue
	status statusMsg
}

type containerMsg struct {
	val    containerMsgValue
	status statusMsg
}

type listUpdateMsg struct {
	val              []list.Item
	title            string
	statusTxt        string
	status           statusMsg
	preSelectedValue string
}
