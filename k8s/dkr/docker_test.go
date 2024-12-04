package dkr

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDeserialization(t *testing.T) {
	date := DockerCompose{
		Version: "1.0",
		Services: map[string]Service{
			"metering": {
				Image:   "metering-image",
				Volumes: []string{"/path/to/volume1", "/path/to/volume2"},
			},
			"otc-container": {
				Image:   "otc-container-image",
				Volumes: []string{"/path/to/volume3"},
			},
		},
		Volumes: map[string]Volume{
			"app-configs": {
				Driver: "local",
				DriverOpts: DriverOpts{
					Type:   "nfs",
					O:      "addr=localhost",
					Device: "/mnt/volume",
				},
			},
			"kube-api-access-tppn9": {
				Driver: "local",
				DriverOpts: DriverOpts{
					Type:   "nfs",
					O:      "addr=localhost",
					Device: "/mnt/volume2",
				},
			},
		},
	}

	yamlData, err := yaml.Marshal(&date)
	if err != nil {
		t.Error("Error while Marshaling", err)
	}

	t.Log(string(yamlData))
}
