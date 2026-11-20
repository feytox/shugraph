package handler

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

type scenarioServiceMock struct {
	ScenarioService
	run func(scenarioID string, delay time.Duration) error
}

func (m scenarioServiceMock) Run(scenarioID string, delay time.Duration) error {
	return m.run(scenarioID, delay)
}

func TestScenarioHandlerRun(t *testing.T) {
	tests := map[string]struct {
		mockRun    func(scenarioID string, delay time.Duration) error
		url        string
		wantStatus int
	}{
		"success": {
			mockRun: func(scenarioID string, delay time.Duration) error {
				assert.Equal(t, "my-scenario", scenarioID)
				assert.Equal(t, 2*time.Second, delay)

				return nil
			},
			url:        "/admin-api/v1/scenarios/my-scenario/run?delay_s=2",
			wantStatus: http.StatusAccepted,
		},
		"scenario not found": {
			mockRun: func(string, time.Duration) error {
				return domain.ErrScenarioNotFound
			},
			url:        "/admin-api/v1/scenarios/unknown-scenario/run",
			wantStatus: http.StatusNotFound,
		},
		"scenario format error": {
			mockRun: func(scenarioID string, delay time.Duration) error {
				return fmt.Errorf("prepare scenario")
			},
			url:        "/admin-api/v1/scenarios/my-broken-scenario/run",
			wantStatus: http.StatusInternalServerError,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			scenarios := scenarioServiceMock{run: tt.mockRun}

			rec := serve(t, Deps{Scenarios: scenarios}, http.MethodPost, tt.url, "")

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
