package sokofile

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Name  string          `yaml:"name"`
	Flows map[string]Flow `yaml:"flows"`
}

type Flow struct {
	Steps    []FlowStep    `yaml:"steps"`
	Schedule *FlowSchedule `yaml:"schedule"`
}

type FlowSchedule struct {
	MinuteValue string `yaml:"minute"`
	HourValue   string `yaml:"hour"`
	DayValue    string `yaml:"day"`
}

func parseValue[T any](value string, convert func(string) (T, error)) []T {
	if value == "*" {
		return nil
	}

	values := strings.Split(value, ",")
	res := make([]T, 0)

	for _, val := range values {
		num, err := convert(strings.TrimSpace(val))
		if err == nil {
			res = append(res, num)
		}
	}

	return res
}

func (s FlowSchedule) Minutes() []int {
	return parseValue(s.MinuteValue, strconv.Atoi)
}

func (s FlowSchedule) Hours() []int {
	return parseValue(s.HourValue, strconv.Atoi)
}

func (s FlowSchedule) Days() []time.Weekday {
	return parseValue(s.DayValue, func(value string) (time.Weekday, error) {
		for i := range 7 {
			if strings.EqualFold(value, time.Weekday(i).String()) {
				return time.Weekday(i), nil
			}
		}
		return 0, errors.New("parse: weekday not found")
	})
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
