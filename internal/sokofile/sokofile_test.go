package sokofile_test

import (
	"slices"
	"testing"
	"time"

	"github.com/fourls/soko/internal/sokofile"
)

func TestScheduleMinutes(t *testing.T) {
	cases := []struct {
		input    string
		expected []int
	}{
		{"*", nil},
		{"25", []int{25}},
		{"1,2,3", []int{1, 2, 3}},
		{"1, 2 , 3", []int{1, 2, 3}},
		{"1,55,31", []int{1, 55, 31}},
		{"1,foo, a bar,,", []int{1}},
	}

	for _, tc := range cases {
		schedule := sokofile.FlowSchedule{
			MinuteValue: tc.input,
		}
		result := schedule.Minutes()
		if !slices.Equal(result, tc.expected) {
			t.Fatalf("got: %v, want: %v", result, tc.expected)
		}
	}
}

func TestScheduleHours(t *testing.T) {
	cases := []struct {
		input    string
		expected []int
	}{
		{"*", nil},
		{"1", []int{1}},
		{"4,5,3", []int{4, 5, 3}},
		{"1, 2 , 3", []int{1, 2, 3}},
		{"1,23,11", []int{1, 23, 11}},
		{"1,foo, a bar,,", []int{1}},
	}

	for _, tc := range cases {
		schedule := sokofile.FlowSchedule{
			HourValue: tc.input,
		}
		result := schedule.Hours()
		if !slices.Equal(result, tc.expected) {
			t.Fatalf("got: %v, want: %v", result, tc.expected)
		}
	}
}

func TestScheduleDays(t *testing.T) {
	cases := []struct {
		input    string
		expected []time.Weekday
	}{
		{"*", nil},
		{"Monday", []time.Weekday{time.Monday}},
		{"Monday  , Tuesday,Saturday", []time.Weekday{time.Monday, time.Tuesday, time.Saturday}},
		{"monday,friday", []time.Weekday{time.Monday, time.Friday}},
		{"mon,fri,sat", []time.Weekday{}},
	}

	for _, tc := range cases {
		schedule := sokofile.FlowSchedule{
			DayValue: tc.input,
		}
		result := schedule.Days()
		if !slices.Equal(result, tc.expected) {
			t.Fatalf("got: %v, want: %v", result, tc.expected)
		}
	}
}
