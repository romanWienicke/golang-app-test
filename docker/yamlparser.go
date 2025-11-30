package docker

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Ports struct {
	Port     string `yaml:"target"`
	Protocol string `yaml:"protocol"`
}

type Service struct {
	Ports []Ports `yaml:"ports"`
}

type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
}

func ParseDockerComposeFile(filename string) (ComposeFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ComposeFile{}, err
	}

	var compose ComposeFile
	err = yaml.Unmarshal(data, &compose)
	if err != nil {
		return ComposeFile{}, err
	}

	return compose, nil
}
