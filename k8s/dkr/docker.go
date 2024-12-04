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
	Image         string
	Volumes       []string
	Environnement map[string]string `yaml:"environment"`
}

type DockerCompose struct {
	Version  string
	Services map[string]Service
	Volumes  map[string]Volume
}
