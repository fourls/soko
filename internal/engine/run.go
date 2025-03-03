package engine

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func runJob(job *Job, report func(func(info *JobInfo))) bool {
	report(func(info *JobInfo) {
		info.State = JobRunning
	})

	for i, step := range job.Steps {
		var output []byte = nil
		var state JobState = JobRunning

		input := strings.Join(step.Args, " ")

		output, err := runStep(&step)
		if err != nil {
			state = JobFailed
			output = []byte(fmt.Sprintf("Step failed with error:\n  %s\n\n%s", err.Error(), output))
		}

		report(func(info *JobInfo) {
			info.CurrentStep = i
			info.Steps[i].Input = input
			info.Steps[i].Output = output
			info.State = state
		})

		if state == JobFailed {
			return false
		}
	}

	report(func(info *JobInfo) {
		info.State = JobSucceeded
	})

	return true
}

func runStep(step *Step) ([]byte, error) {
	if len(step.Args) == 0 {
		return nil, errors.New("Step is empty")
	}

	cmd := exec.Command(step.Args[0], step.Args[1:]...)
	return cmd.CombinedOutput()
}
