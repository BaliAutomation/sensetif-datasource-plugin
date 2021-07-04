package model

type Query struct {
	*SensorRef
	Alias string
}

type SensorRef struct {
	Project   string
	Subsystem string
	Datapoint string
}

func (q *Query) String() string {
	result := q.Project
	if len(q.Subsystem) > 0 {
		result += "/" + q.Subsystem
	}

	if len(q.Datapoint) > 0 {
		result += "/" + q.Datapoint
	}

	return result
}
