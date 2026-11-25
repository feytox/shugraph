package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

const scenarioTag = "Scenarios"

type ScenarioService interface {
	Run(scenarioID string, delay time.Duration) error
	ListNames() ([]string, error)
}

type scenarioHandlers struct {
	scenario ScenarioService
	log      *slog.Logger
}

func registerScenarios(api huma.API, h *scenarioHandlers) {
	huma.Register(api, huma.Operation{
		OperationID:   "run-scenario",
		Summary:       "Run a scenario in graph-api",
		Description:   "Find the scenario ID using List scenarios. A 202 Accepted response means the scenario was accepted for background playback. Check graph-api to see the results.",
		Method:        http.MethodPost,
		Path:          "/admin-api/v1/scenarios/{scenario_id}/run",
		Tags:          []string{scenarioTag},
		DefaultStatus: http.StatusAccepted,
		Errors:        []int{http.StatusNotFound, http.StatusInternalServerError},
	}, h.run)

	huma.Register(api, huma.Operation{
		OperationID:   "list-scenarios",
		Summary:       "List scenarios",
		Description:   "Return the available scenario IDs. Use an ID from this list to run a scenario.",
		Method:        http.MethodGet,
		Path:          "/admin-api/v1/scenarios",
		Tags:          []string{scenarioTag},
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusInternalServerError},
	}, h.list)
}

type RunScenarioInput struct {
	ScenarioID   string `path:"scenario_id" doc:"Scenario ID from List scenarios (for example, device_farm)." example:"device_farm"`
	DelaySeconds int    `query:"delay_s" minimum:"0" default:"1" maximum:"5" doc:"Delay in seconds before sending each event. Set to 0 to send events without waiting."`
}

type ListScenarioOutput struct {
	Body []string `nullable:"false"`
}

func (h scenarioHandlers) run(ctx context.Context, i *RunScenarioInput) (*struct{}, error) {
	err := h.scenario.Run(i.ScenarioID, time.Duration(i.DelaySeconds)*time.Second)
	if err != nil {
		if errors.Is(err, domain.ErrScenarioNotFound) {
			return nil, huma.Error404NotFound("Scenario not found")
		}

		return nil, problem(ctx, h.log, err)
	}

	return nil, nil
}

func (h scenarioHandlers) list(ctx context.Context, _ *struct{}) (*ListScenarioOutput, error) {
	names, err := h.scenario.ListNames()
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	return &ListScenarioOutput{Body: names}, nil
}
