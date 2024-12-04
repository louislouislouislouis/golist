package k8s

import (
	"context"
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

func TestGetVolumes(t *testing.T) {
	k8s := GetService()
	k8s.PodToContainer("kiwios-cloud-metering", "metering-dc5654c89-hxqjq", context.TODO())
}
