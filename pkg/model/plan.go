package model

import "time"

type PlanLimits struct {
	MaxProjects      int16        `json:"maxprojects"`
	MaxCollaborators int16        `json:"maxcollaborators"`
	MaxDatapoints    int32        `json:"maxdatapoints"`
	MaxStorage       TimeToLive   `json:"maxstorage"`
	MinPollInterval  PollInterval `json:"minpollinterval"`
}

type PlanSettings struct {
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Private     bool       `json:"private"`
	Active      bool       `json:"active"`
	Limits      PlanLimits `json:"limits"`
	Start       time.Time  `json:"start"`
	End         time.Time  `json:"end"`
}
