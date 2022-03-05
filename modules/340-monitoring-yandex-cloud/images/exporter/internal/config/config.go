package config

type Instance struct {
	Token             string `yaml:"token"`
	ServiceType       string `yaml:"serviceType"`
	SessionTimeoutSec string `yaml:"sessionTimeoutSec"`
}

type Server struct {
	Path string `yaml:"path"`
	Port string `yaml:"port"`
	Url  string `yaml:"url"`
}
