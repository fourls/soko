package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fourls/soko/internal/engine"
	"github.com/fourls/soko/internal/sokofile"
	"github.com/gorilla/mux"
)

type JobDto struct {
	JobId  string          `json:"id"`
	FlowId string          `json:"flow"`
	State  string          `json:"state"`
	Output []StepResultDto `json:"output"`
}

type StepResultDto struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func jobInfoToDto(id engine.JobId, info *engine.JobInfo) JobDto {
	state := "unknown"

	switch info.State {
	case engine.JobPending:
		state = "pending"
	case engine.JobRunning:
		state = "running"
	case engine.JobSucceeded:
		state = "succeeded"
	case engine.JobFailed:
		state = "failed"
	}

	output := make([]StepResultDto, len(info.Steps))
	for i, step := range info.Steps {
		// todo sanitize
		output[i] = StepResultDto{
			Input:  step.Input,
			Output: string(step.Output),
		}
	}

	return JobDto{
		JobId:  string(id),
		FlowId: string(info.FlowId),
		State:  state,
		Output: output,
	}
}

func sokofileToFlows(project *sokofile.Project) map[engine.FlowId]engine.Flow {
	flows := make(map[engine.FlowId]engine.Flow, len(project.Flows))

	i := 0
	for key, value := range project.Flows {
		steps := make([]engine.Step, len(value.Steps))

		for j, step := range value.Steps {
			steps[j] = engine.Step{
				Args: step.Cmd,
			}
		}

		id := engine.FlowId(project.Name + "." + key)

		var schedule *engine.FlowSchedule = nil
		if value.Schedule != nil {
			schedule = &engine.FlowSchedule{
				Minutes: value.Schedule.Minutes(),
				Hours:   value.Schedule.Hours(),
				Days:    value.Schedule.Days(),
			}
		}

		flows[id] = engine.Flow{
			Id:       id,
			Steps:    steps,
			Schedule: schedule,
		}
		i++
	}

	return flows
}

func main() {
	jobEngine := engine.New()

	project, err := sokofile.Parse("soko.yml")
	if err != nil {
		panic(err)
	}
	jobEngine.Flows = sokofileToFlows(project)

	jobEngine.StartWorkers()
	defer jobEngine.StopWorkers()

	router := mux.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world!")
	})

	router.HandleFunc("/flows/{id}/run", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flowId := engine.FlowId(vars["id"])
		found, jobId := jobEngine.StartJob(flowId)

		if found {
			json.NewEncoder(w).Encode(JobDto{
				JobId:  string(jobId),
				FlowId: string(flowId),
				State:  "pending",
				Output: make([]StepResultDto, 0),
			})
		} else {
			http.Error(w, "Flow not found", 404)
		}
	}).Methods("POST")

	router.HandleFunc("/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := engine.JobId(vars["id"])

		info := jobEngine.GetJob(id)

		if info == nil {
			http.Error(w, "Job not found", 404)
			return
		}

		json.NewEncoder(w).Encode(jobInfoToDto(id, info))
	}).Methods("GET")

	http.ListenAndServe(":8000", router)
}
