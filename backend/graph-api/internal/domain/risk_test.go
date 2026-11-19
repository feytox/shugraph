package domain

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var riskMDRules = RiskRules{
	Levels: RiskSteps[RiskLevel]{
		{From: 0, Value: RiskLevelLow},
		{From: 60, Value: RiskLevelMedium},
		{From: 75, Value: RiskLevelHigh},
	},
	SpeedBySpanMinutes: RiskSteps[int]{
		{From: 0, Value: 20},
		{From: 11, Value: 15},
		{From: 31, Value: 10},
		{From: 61, Value: 5},
		{From: 181, Value: 0},
	},
	Farm: FarmRiskRules{
		SchemePoints: 50,
		ClientsPoints: RiskSteps[int]{
			{From: 3, Value: 0},
			{From: 4, Value: 5},
			{From: 5, Value: 10},
			{From: 6, Value: 15},
		},
		CollectorPoints: 10,
	},
}

func statsWithSpan(span time.Duration) GroupStats {
	first := time.Date(2026, 11, 12, 10, 1, 0, 0, time.UTC)
	return GroupStats{FirstEventAt: first, LastEventAt: first.Add(span)}
}

func farmParts(t *testing.T, risk Risk) farmRiskParts {
	t.Helper()

	var parts map[GroupScheme]farmRiskParts
	require.NoError(t, json.Unmarshal(risk.Parts, &parts))
	require.Len(t, parts, 1)

	return parts[GroupSchemeFarm]
}

func TestRiskRules_ScoreFarm(t *testing.T) {
	tests := []struct {
		name      string
		rules     RiskRules
		span      time.Duration
		facts     FarmFacts
		wantScore int
		wantLevel RiskLevel
		wantParts string
	}{
		{
			name:      "scenario device_farm",
			rules:     riskMDRules,
			span:      28 * time.Minute,
			facts:     FarmFacts{ClientsCount: 3, HasCollector: true},
			wantScore: 75,
			wantLevel: RiskLevelHigh,
			wantParts: `{"farm": {"scheme": 50, "speed": 15, "clients": 0, "collector": 10}}`,
		},
		{
			name:      "scenario farm_and_address",
			rules:     riskMDRules,
			span:      45 * time.Minute,
			facts:     FarmFacts{ClientsCount: 3, HasCollector: true},
			wantScore: 70,
			wantLevel: RiskLevelMedium,
			wantParts: `{"farm": {"scheme": 50, "speed": 10, "clients": 0, "collector": 10}}`,
		},
		{
			name:      "no collector",
			rules:     riskMDRules,
			span:      4 * time.Hour,
			facts:     FarmFacts{ClientsCount: 3},
			wantScore: 50,
			wantLevel: RiskLevelLow,
			wantParts: `{"farm": {"scheme": 50, "speed": 0, "clients": 0, "collector": 0}}`,
		},
		{
			name: "score is capped at 100",
			rules: func() RiskRules {
				rules := riskMDRules
				rules.Farm.SchemePoints = 90

				return rules
			}(),
			span:      5 * time.Minute,
			facts:     FarmFacts{ClientsCount: 6, HasCollector: true},
			wantScore: 100,
			wantLevel: RiskLevelHigh,
			wantParts: `{"farm": {"scheme": 90, "speed": 20, "clients": 15, "collector": 10}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := tt.rules.ScoreFarm(statsWithSpan(tt.span), tt.facts)

			require.Equal(t, tt.wantScore, risk.Score)
			require.Equal(t, tt.wantLevel, risk.Level)
			require.JSONEq(t, tt.wantParts, string(risk.Parts))
		})
	}
}

func TestRiskRules_ScoreFarmSpeed(t *testing.T) {
	tests := []struct {
		span      time.Duration
		wantSpeed int
	}{
		{0, 20},
		{10 * time.Minute, 20},
		{10*time.Minute + 59*time.Second, 20},
		{11 * time.Minute, 15},
		{30 * time.Minute, 15},
		{31 * time.Minute, 10},
		{60 * time.Minute, 10},
		{61 * time.Minute, 5},
		{180 * time.Minute, 5},
		{181 * time.Minute, 0},
		{20 * time.Hour, 0},
	}

	for _, tt := range tests {
		t.Run(tt.span.String(), func(t *testing.T) {
			risk := riskMDRules.ScoreFarm(statsWithSpan(tt.span), FarmFacts{ClientsCount: 3})

			require.Equal(t, tt.wantSpeed, farmParts(t, risk).Speed)
		})
	}
}

func TestRiskRules_ScoreFarmClients(t *testing.T) {
	tests := []struct {
		clients     int
		wantClients int
	}{
		{3, 0},
		{4, 5},
		{5, 10},
		{6, 15},
		{9, 15},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d clients", tt.clients), func(t *testing.T) {
			risk := riskMDRules.ScoreFarm(statsWithSpan(time.Hour), FarmFacts{ClientsCount: tt.clients})

			require.Equal(t, tt.wantClients, farmParts(t, risk).Clients)
		})
	}
}

func TestRiskSteps_Levels(t *testing.T) {
	tests := []struct {
		score     int
		wantLevel RiskLevel
	}{
		{0, RiskLevelLow},
		{59, RiskLevelLow},
		{60, RiskLevelMedium},
		{74, RiskLevelMedium},
		{75, RiskLevelHigh},
		{100, RiskLevelHigh},
	}

	for _, tt := range tests {
		require.Equal(t, tt.wantLevel, riskMDRules.Levels.At(tt.score), "score %d", tt.score)
	}
}
