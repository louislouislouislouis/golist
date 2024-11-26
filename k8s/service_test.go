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

func TestGetEnvValues(t *testing.T) {
	k8s := GetService()
	k8s.GetEnvFromContainer(
		"kiwios-cloud-metering",
		"metering-5668bbd986-lhrbv",
		"",
		context.TODO(),
	)
}

func TestGetVolumes(t *testing.T) {
	k8s := GetService()
	k8s.listVolumes("kiwios-cloud-metering", "metering-5668bbd986-lhrbv", context.TODO())
}
