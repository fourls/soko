package html

import (
	"embed"
	"html/template"
	"io"
)

//go:embed *
var files embed.FS

func parse(file string) *template.Template {
	return template.Must(template.New("layout.html").ParseFS(files, "layout.html", file))
}

var (
	dashboardTemplate = parse("dashboard.html")
)

type Flow struct {
	Id       string
	Name     string
	Schedule string
	Jobs     []*Job
}

type Job struct {
	Id     string
	Flow   *Flow
	State  string
	FlowId string
}

type DashboardParams struct {
	Flows map[string]Flow
	Jobs  map[string]Job
}

func Dashboard(w io.Writer, p DashboardParams) error {
	return dashboardTemplate.Execute(w, p)
}
