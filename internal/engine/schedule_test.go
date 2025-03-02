package engine_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/fourls/soko/internal/engine"
)

func TestScheduleMatches(t *testing.T) {
	date := time.Date(2024, 02, 20, 14, 03, 25, 0, time.UTC)

	cases := []struct {
		schedule engine.FlowSchedule
		time     time.Time
		expected bool
	}{
		{
			engine.FlowSchedule{},
			date,
			true,
		},
		{
			engine.FlowSchedule{
				Minutes: []int{date.Minute(), 40},
			},
			date,
			true,
		},
		{
			engine.FlowSchedule{
				Minutes: []int{date.Minute() + 1},
			},
			date,
			false,
		},
		{
			engine.FlowSchedule{
				Hours: []int{date.Hour()},
			},
			date,
			true,
		},
		{
			engine.FlowSchedule{
				Hours: []int{date.Hour() - 1},
			},
			date,
			false,
		},
		{
			engine.FlowSchedule{
				Hours: []int{date.Hour() - 12},
			},
			date,
			false,
		},
		{
			engine.FlowSchedule{
				Minutes: []int{date.Minute()},
				Hours:   []int{date.Hour()},
				Days:    []time.Weekday{date.Weekday()},
			},
			date,
			true,
		},
		{
			engine.FlowSchedule{
				Minutes: []int{date.Minute()},
				Hours:   []int{date.Hour() + 2},
				Days:    []time.Weekday{date.Weekday()},
			},
			date,
			false,
		},
		{
			engine.FlowSchedule{
				Minutes: []int{date.Minute() + 5},
				Hours:   []int{date.Hour() + 2},
				Days:    []time.Weekday{date.Weekday() - 1},
			},
			date,
			false,
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			matches := tc.schedule.Matches(tc.time)
			if matches != tc.expected {
				t.Fatalf("[%d] got: %v, expected: %v", i, matches, tc.expected)
			}
		})
	}
}
