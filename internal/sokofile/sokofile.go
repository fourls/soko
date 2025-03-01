package sokofile

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Name  string          `yaml:"name"`
	Flows map[string]Flow `yaml:"flows"`
}

type Flow struct {
	Steps []FlowStep `yaml:"steps"`
}

type FlowStep struct {
	Cmd []string `yaml:"cmd"`
}

func Parse(file string) (*Project, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	var contents Project
	err = yaml.NewDecoder(f).Decode(&contents)
	return &contents, err
}
