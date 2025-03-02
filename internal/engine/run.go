package engine

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func runJob(job *Job, updateChan chan jobUpdate) bool {
	updateChan <- miscJobStateUpdate{id: job.Id, jobState: JobRunning}

	for i, step := range job.Steps {
		var output []byte = nil
		var state JobState = JobRunning

		output, err := runStep(&step)
		if err != nil {
			state = JobFailed
			output = []byte(fmt.Sprintf("Step failed with error:\n  %s\n\n%s", err.Error(), output))
		}

		updateChan <- jobStepUpdateImpl{
			id:         job.Id,
			stepIndex:  i,
			jobState:   state,
			stepInput:  strings.Join(step.Args, " "),
			stepOutput: output,
		}

		if state == JobFailed {
			return false
		}
	}

	updateChan <- miscJobStateUpdate{id: job.Id, jobState: JobSucceeded}
	return true
}

func runStep(step *Step) ([]byte, error) {
	if len(step.Args) == 0 {
		return nil, errors.New("Step is empty")
	}

	cmd := exec.Command(step.Args[0], step.Args[1:]...)
	return cmd.CombinedOutput()
}
