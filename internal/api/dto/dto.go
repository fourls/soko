package dto

import "github.com/fourls/soko/internal/engine"

type Job struct {
	JobId  string       `json:"id"`
	FlowId string       `json:"flow"`
	State  string       `json:"state"`
	Output []StepResult `json:"output"`
}

type StepResult struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func FromJobInfo(id engine.JobId, info *engine.JobInfo) Job {
	output := make([]StepResult, len(info.Steps))
	for i, step := range info.Steps {
		// todo sanitize
		output[i] = StepResult{
			Input:  step.Input,
			Output: string(step.Output),
		}
	}

	return Job{
		JobId:  string(id),
		FlowId: string(info.FlowId),
		State:  info.State.String(),
		Output: output,
	}
}
