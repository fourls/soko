package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fourls/soko/internal/api/dto"
	"github.com/fourls/soko/internal/engine"
	"github.com/gorilla/mux"
)

func ConfigureRouter(router *mux.Router, jobEngine *engine.JobEngine) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			Message string `json:"message"`
		}{"pong"})
	})

	router.HandleFunc("/flows/{id}/run", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		flowId := engine.FlowId(vars["id"])
		found, jobId := jobEngine.StartJob(flowId)

		if found {
			json.NewEncoder(w).Encode(dto.Job{
				JobId:  string(jobId),
				FlowId: string(flowId),
				State:  "pending",
				Output: make([]dto.StepResult, 0),
			})
		} else {
			http.Error(w, "Flow not found", 404)
		}
	}).Methods("POST")

	router.HandleFunc("/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := engine.JobId(vars["id"])

		info, ok := jobEngine.GetJob(id)
		if !ok {
			http.Error(w, "Job not found", 404)
			return
		}

		json.NewEncoder(w).Encode(dto.FromJobInfo(id, &info))
	}).Methods("GET")

	log.Print("Configured API routes")
}
