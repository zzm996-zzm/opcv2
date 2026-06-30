package dashboard

import "errors"

var ErrServiceNotReady = errors.New("dashboard service is not configured")

type Summary struct {
	Metrics  []Metric             `json:"metrics"`
	Projects []ProjectOpportunity `json:"projects"`
	Trend    []TrendPoint         `json:"trend"`
	Pipeline []PipelineStage      `json:"pipeline"`
	Alerts   []Alert              `json:"alerts"`
	Actions  []Action             `json:"actions"`
}

type Metric struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Change string `json:"change"`
}

type ProjectOpportunity struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Leads string `json:"leads"`
	Stage string `json:"stage"`
}

type TrendPoint struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type PipelineStage struct {
	Stage   string `json:"stage"`
	Count   string `json:"count"`
	Percent string `json:"percent"`
}

type Alert struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type Action struct {
	Time  string `json:"time"`
	Title string `json:"title"`
}
