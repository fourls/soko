package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fourls/soko/internal/schedule"
	"github.com/gorilla/mux"
)

type JobsGetResponse struct {
	JobId string
	State string
}

type FlowsRunResponse struct {
	JobId string
}

func main() {
	scheduler := schedule.New()
	scheduler.Flows = map[schedule.FlowId]schedule.Flow{
		"foo": {
			Id: "foo",
			Steps: []schedule.Step{
				{
					Args: []string{"touch", "empty.txt"},
				},
				{
					Args: []string{"bash", "-c", "echo 'Hello world!' > hello.txt"},
				},
			},
		},
	}

	go scheduler.ProcessJobsWorker()

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world!")
	})

	router.HandleFunc("/flows/{id}/run", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flowId := schedule.FlowId(vars["id"])
		found, jobId := scheduler.StartJob(flowId)

		if found {
			json.NewEncoder(w).Encode(FlowsRunResponse{JobId: string(jobId)})
		} else {
			http.Error(w, "Flow not found", 404)
		}
	})

	router.HandleFunc("/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := schedule.JobId(vars["id"])

		scheduler.ProcessUpdates()
		state, ok := scheduler.Jobs[id]

		if !ok {
			http.Error(w, "Job not found", 404)
			return
		}

		resp := JobsGetResponse{JobId: string(id)}

		switch state {
		case schedule.JobPending:
			resp.State = "pending"
		case schedule.JobRunning:
			resp.State = "running"
		case schedule.JobSucceeded:
			resp.State = "succeeded"
		case schedule.JobFailed:
			resp.State = "failed"
		default:
			resp.State = fmt.Sprintf("unknown (%d)", state)
		}

		json.NewEncoder(w).Encode(resp)
	})

	http.ListenAndServe(":8000", router)
}
