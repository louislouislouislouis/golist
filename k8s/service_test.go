package k8s

import (
	"testing"
)

func TestListNamespace(t *testing.T) {
	t.Log("Testinqg the listService Functionality")
}

func TestGetContainer(t *testing.T) {
	k8s := GetService()
	liste, _ := k8s.ListNamespace()
	t.Log(liste.Items)
}
