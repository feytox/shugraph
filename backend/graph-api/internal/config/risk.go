package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

type Risk struct {
	LevelByScore             levelByScore      `env:"RISK_LEVEL_BY_SCORE" envDefault:"0:low,60:medium,75:high"`
	SpeedPointsBySpanMinutes pointsByThreshold `env:"RISK_SPEED_POINTS_BY_SPAN_MINUTES" envDefault:"0:20,11:15,31:10,61:5,181:0"`
	FarmSchemePoints         int               `env:"RISK_FARM_SCHEME_POINTS" envDefault:"50"`
	FarmPointsByClients      pointsByThreshold `env:"RISK_FARM_POINTS_BY_CLIENTS" envDefault:"3:0,4:5,5:10,6:15"`
	FarmCollectorPoints      int               `env:"RISK_FARM_COLLECTOR_POINTS" envDefault:"10"`
}

func (r Risk) Rules() domain.RiskRules {
	return domain.RiskRules{
		Levels:             domain.RiskSteps[domain.RiskLevel](r.LevelByScore),
		SpeedBySpanMinutes: domain.RiskSteps[int](r.SpeedPointsBySpanMinutes),
		Farm: domain.FarmRiskRules{
			SchemePoints:    r.FarmSchemePoints,
			ClientsPoints:   domain.RiskSteps[int](r.FarmPointsByClients),
			CollectorPoints: r.FarmCollectorPoints,
		},
	}
}

type pointsByThreshold domain.RiskSteps[int]

func (t *pointsByThreshold) UnmarshalText(text []byte) error {
	steps, err := parseSteps(string(text), strconv.Atoi)
	if err != nil {
		return err
	}

	*t = pointsByThreshold(steps)

	return nil
}

type levelByScore domain.RiskSteps[domain.RiskLevel]

var errLevelsMustStartAtZero = errors.New("risk levels must start at score 0")

func (t *levelByScore) UnmarshalText(text []byte) error {
	steps, err := parseSteps(string(text), parseRiskLevel)
	if err != nil {
		return err
	}

	if steps[0].From != 0 {
		return errLevelsMustStartAtZero
	}

	*t = levelByScore(steps)

	return nil
}

var riskLevels = []domain.RiskLevel{domain.RiskLevelLow, domain.RiskLevelMedium, domain.RiskLevelHigh}

func parseRiskLevel(text string) (domain.RiskLevel, error) {
	level := domain.RiskLevel(text)
	if !slices.Contains(riskLevels, level) {
		return "", fmt.Errorf("unknown risk level %q, want one of %v", text, riskLevels)
	}

	return level, nil
}

func parseSteps[T any](text string, parseValue func(string) (T, error)) (domain.RiskSteps[T], error) {
	pairs := strings.Split(text, ",")
	steps := make(domain.RiskSteps[T], 0, len(pairs))

	for _, pair := range pairs {
		step, err := parseStep(strings.TrimSpace(pair), parseValue)
		if err != nil {
			return nil, err
		}

		if len(steps) > 0 && step.From <= steps[len(steps)-1].From {
			return nil, fmt.Errorf("risk step %q: thresholds must increase, %d goes after %d", pair, step.From, steps[len(steps)-1].From)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func parseStep[T any](pair string, parseValue func(string) (T, error)) (domain.RiskStep[T], error) {
	threshold, value, ok := strings.Cut(pair, ":")
	if !ok {
		return domain.RiskStep[T]{}, fmt.Errorf("risk step %q: want threshold:value", pair)
	}

	from, err := strconv.Atoi(threshold)
	if err != nil {
		return domain.RiskStep[T]{}, fmt.Errorf("risk step %q: threshold: %w", pair, err)
	}

	parsed, err := parseValue(value)
	if err != nil {
		return domain.RiskStep[T]{}, fmt.Errorf("risk step %q: value: %w", pair, err)
	}

	return domain.RiskStep[T]{From: from, Value: parsed}, nil
}
