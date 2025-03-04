package engine

import (
	"fmt"
	"slices"
	"strings"
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

func (s FlowSchedule) String() string {
	days := ""
	minutes := "every minute"

	if s.Days != nil {
		combined := make([]string, len(s.Days))
		for i, day := range s.Days {
			combined[i] = day.String()
		}

		days = "on " + strings.Join(combined, ", ")
	}

	if s.Hours != nil {
		minutesCount := 1
		if s.Minutes != nil {
			minutesCount = len(s.Minutes)
		}

		combined := make([]string, len(s.Hours)*minutesCount)
		for i, hour := range s.Hours {
			if s.Minutes == nil {
				combined[i] = fmt.Sprintf("every minute from %d:00 to %d:59", hour, hour)
			}

			for j, minute := range s.Minutes {
				combined[i*minutesCount+j] = fmt.Sprintf("%d:%d", hour, minute)
			}
		}

		minutes = strings.Join(combined, ", ")
	}

	if days == "" {
		return minutes
	} else {
		return fmt.Sprintf("%s %s", minutes, days)
	}
}
