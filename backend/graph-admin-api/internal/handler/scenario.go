package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

const scenarioTag = "Scenarios"

type ScenarioService interface {
	Run(ctx context.Context, scenarioID string, delay int) error
	ListNames() ([]string, error)
}

type scenarioHandlers struct {
	scenario ScenarioService
	log      *slog.Logger
}

func registerService(api huma.API, h *scenarioHandlers) {
	huma.Register(api, huma.Operation{
		OperationID:   "run-scenario",
		Summary:       "Run a scenario in graph-api",
		Method:        http.MethodPost,
		Path:          "/admin-api/v1/scenario/{scenario_id}/run",
		Tags:          []string{scenarioTag},
		DefaultStatus: http.StatusAccepted,
		Errors:        []int{http.StatusNotFound},
	}, h.run)

	huma.Register(api, huma.Operation{
		OperationID:   "list-scenarios",
		Summary:       "List a scenario",
		Method:        http.MethodGet,
		Path:          "/admin-api/v1/scenario",
		Tags:          []string{scenarioTag},
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusInternalServerError},
	}, h.list)
}

type RunScenarioInput struct {
	ScenarioID   string `path:"scenario_id" validate:"required"`
	DelaySeconds int    `query:"delay_s" minimum:"0" default:"1" maximum:"5"`
}

type ListScenarioOutput struct {
	Body []string
}

func (h scenarioHandlers) run(ctx context.Context, i *RunScenarioInput) (*struct{}, error) {
	err := h.scenario.Run(ctx, i.ScenarioID, i.DelaySeconds)
	if errors.Is(err, domain.ErrScenarioNotFound) {
		return nil, huma.Error404NotFound("Scenario not found")
	}
	return nil, problem(ctx, h.log, err)
}

func (h scenarioHandlers) list(ctx context.Context, _ *struct{}) (*ListScenarioOutput, error) {
	names, err := h.scenario.ListNames()
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}
	return &ListScenarioOutput{Body: names}, nil
}
