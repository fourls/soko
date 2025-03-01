package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fourls/soko/internal/schedule"
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

func jobInfoToDto(id schedule.JobId, info *schedule.JobInfo) JobDto {
	state := "unknown"

	switch info.State {
	case schedule.JobPending:
		state = "pending"
	case schedule.JobRunning:
		state = "running"
	case schedule.JobSucceeded:
		state = "succeeded"
	case schedule.JobFailed:
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

func sokofileToFlows(project *sokofile.Project) map[schedule.FlowId]schedule.Flow {
	flows := make(map[schedule.FlowId]schedule.Flow, len(project.Flows))

	i := 0
	for key, value := range project.Flows {
		steps := make([]schedule.Step, len(value.Steps))

		for j, step := range value.Steps {
			steps[j] = schedule.Step{
				Args: step.Cmd,
			}
		}

		id := schedule.FlowId(project.Name + "." + key)

		flows[id] = schedule.Flow{
			Id:    id,
			Steps: steps,
		}
		i++
	}

	return flows
}

func main() {
	scheduler := schedule.New()

	project, err := sokofile.Parse("soko.yml")
	if err != nil {
		panic(err)
	}
	scheduler.Flows = sokofileToFlows(project)

	scheduler.StartWorkers()
	defer scheduler.StopWorkers()

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
		flowId := schedule.FlowId(vars["id"])
		found, jobId := scheduler.StartJob(flowId)

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
		id := schedule.JobId(vars["id"])

		info := scheduler.GetJob(id)

		if info == nil {
			http.Error(w, "Job not found", 404)
			return
		}

		json.NewEncoder(w).Encode(jobInfoToDto(id, info))
	}).Methods("GET")

	http.ListenAndServe(":8000", router)
}
