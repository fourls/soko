package main

import (
	"log"
	"net/http"

	"github.com/fourls/soko/internal/api"
	"github.com/fourls/soko/internal/engine"
	"github.com/fourls/soko/internal/sokofile"
	"github.com/fourls/soko/internal/web"
	"github.com/gorilla/mux"
)

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
	defer jobEngine.Close()

	// todo support loading projects
	project, err := sokofile.Parse("soko.yml")
	if err != nil {
		panic(err)
	}

	for id, flow := range sokofileToFlows(project) {
		jobEngine.Flows.Create(id, flow)
	}

	log.Print("Configured job engine")

	router := mux.NewRouter()

	apiRouter := router.NewRoute().PathPrefix("/api/").Subrouter()
	api.ConfigureRouter(apiRouter, &jobEngine)
	webRouter := router.NewRoute().Subrouter()
	web.ConfigureRouter(webRouter, &jobEngine)

	log.Print("Serving...")
	http.ListenAndServe(":8000", router)
}
