package dkr

type DriverOpts struct {
	Type   string `yaml:"type"`
	O      string `yaml:"o"`
	Device string `yaml:"device"`
}

type Volume struct {
	Driver     string     `yaml:"driver"`
	DriverOpts DriverOpts `yaml:"driver_opts"`
}

type Service struct {
	NetworkMode   string   `yaml:"network_mode"`
	DependsOn     []string `yaml:"depends_on,omitempty"`
	Image         string
	Volumes       []string
	Environnement map[string]string `yaml:"environment,omitempty"`
	Command       []string          `yaml:"command,omitempty"`
}

type DockerCompose struct {
	Version  string
	Services map[string]Service
	Volumes  map[string]Volume
}
