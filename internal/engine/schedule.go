package engine

import (
	"slices"
	"time"
)

type FlowSchedule struct {
	Minutes []int
	Hours   []int
	Days    []time.Weekday
}

func (s FlowSchedule) Matches(time time.Time) bool {
	return (s.Days == nil || slices.Contains(s.Days, time.Weekday())) &&
		(s.Hours == nil || slices.Contains(s.Hours, time.Hour())) &&
		(s.Minutes == nil || slices.Contains(s.Minutes, time.Minute()))
}
