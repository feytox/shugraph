package config

import (
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/require"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
)

func TestRisk_DefaultsMatchRiskMD(t *testing.T) {
	risk, err := env.ParseAs[Risk]()
	require.NoError(t, err)

	require.Equal(t, domain.RiskRules{
		Levels: domain.RiskSteps[domain.RiskLevel]{
			{From: 0, Value: domain.RiskLevelLow},
			{From: 60, Value: domain.RiskLevelMedium},
			{From: 75, Value: domain.RiskLevelHigh},
		},
		SpeedBySpanMinutes: domain.RiskSteps[int]{
			{From: 0, Value: 20},
			{From: 11, Value: 15},
			{From: 31, Value: 10},
			{From: 61, Value: 5},
			{From: 181, Value: 0},
		},
		Farm: domain.FarmRiskRules{
			SchemePoints: 50,
			ClientsPoints: domain.RiskSteps[int]{
				{From: 3, Value: 0},
				{From: 4, Value: 5},
				{From: 5, Value: 10},
				{From: 6, Value: 15},
			},
			CollectorPoints: 10,
		},
	}, risk.Rules())
}

func TestRisk_OverriddenFromEnv(t *testing.T) {
	t.Setenv("RISK_FARM_POINTS_BY_CLIENTS", " 3:0, 7:20 ")

	risk, err := env.ParseAs[Risk]()
	require.NoError(t, err)

	require.Equal(t, domain.RiskSteps[int]{{From: 3, Value: 0}, {From: 7, Value: 20}}, risk.Rules().Farm.ClientsPoints)
}

func TestRisk_RejectsBrokenSteps(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"trailing comma", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "0:20,"},
		{"no colon", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "0-20"},
		{"threshold is not a number", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "zero:20"},
		{"points are not a number", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "0:twenty"},
		{"repeated threshold", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "0:20,0:15"},
		{"decreasing thresholds", "RISK_SPEED_POINTS_BY_SPAN_MINUTES", "11:15,0:20"},
		{"unknown level", "RISK_LEVEL_BY_SCORE", "0:low,60:mid,75:high"},
		{"levels do not start at zero", "RISK_LEVEL_BY_SCORE", "10:low,60:medium,75:high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)

			_, err := env.ParseAs[Risk]()
			require.Error(t, err)
		})
	}
}
