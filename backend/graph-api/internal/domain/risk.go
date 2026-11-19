package domain

import "encoding/json"

const maxRiskScore = 100

type RiskStep[T any] struct {
	From  int
	Value T
}

type RiskSteps[T any] []RiskStep[T]

func (s RiskSteps[T]) At(n int) T {
	var value T

	for _, step := range s {
		if n < step.From {
			break
		}

		value = step.Value
	}

	return value
}

type RiskRules struct {
	Levels             RiskSteps[RiskLevel]
	SpeedBySpanMinutes RiskSteps[int]
	Farm               FarmRiskRules
}

type FarmRiskRules struct {
	SchemePoints    int
	ClientsPoints   RiskSteps[int]
	CollectorPoints int
}

type FarmFacts struct {
	ClientsCount int
	HasCollector bool
}

type farmRiskParts struct {
	Scheme    int `json:"scheme"`
	Speed     int `json:"speed"`
	Clients   int `json:"clients"`
	Collector int `json:"collector"`
}

func (r RiskRules) ScoreFarm(stats GroupStats, facts FarmFacts) Risk {
	parts := farmRiskParts{
		Scheme:  r.Farm.SchemePoints,
		Speed:   r.SpeedBySpanMinutes.At(stats.SpanMinutes()),
		Clients: r.Farm.ClientsPoints.At(facts.ClientsCount),
	}

	if facts.HasCollector {
		parts.Collector = r.Farm.CollectorPoints
	}

	score := min(maxRiskScore, parts.Scheme+parts.Speed+parts.Clients+parts.Collector)

	return Risk{
		Score: score,
		Parts: mustJSON(map[GroupScheme]farmRiskParts{GroupSchemeFarm: parts}),
		Level: r.Levels.At(score),
	}
}

func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}
