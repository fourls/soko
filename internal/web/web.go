package web

import (
	"log"
	"net/http"

	"github.com/fourls/soko/internal/engine"
	"github.com/fourls/soko/internal/web/html"
	"github.com/gorilla/mux"
)

func ConfigureRouter(router *mux.Router, jobEngine *engine.JobEngine) {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		engineFlows := jobEngine.Flows.Snapshot()
		templateFlows := make(map[string]html.Flow, len(engineFlows))
		for id, flow := range engineFlows {
			templateFlows[string(id)] = html.Flow{
				Id:       string(id),
				Schedule: flow.Schedule.String(),
			}
		}

		engineJobs := jobEngine.Jobs.Snapshot()
		templateJobs := make(map[string]html.Job, len(engineJobs))
		for id, job := range engineJobs {
			templateJobs[string(id)] = html.Job{
				Id:     string(id),
				State:  job.State.String(),
				FlowId: string(job.FlowId),
			}
		}

		html.Dashboard(w, html.DashboardParams{
			Flows: templateFlows,
			Jobs:  templateJobs,
		})
	})
	log.Println("Configured web routes")
}
