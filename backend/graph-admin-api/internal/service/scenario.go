package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi/oapi"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/logattr"
)

type eventAPI interface {
	LoginEvent(ctx context.Context, eventID string, payload oapi.LoginPayload) error
	TransferEvent(ctx context.Context, eventID string, payload oapi.TransferPayload) error
}

type scenarioRepository interface {
	GetByName(name string) (*domain.Scenario, error)
	ListNames() ([]string, error)
}

type scenarioEvent struct {
	EventID string `json:"event_id"`
	Type    string `json:"type"`
}

type ScenarioService struct {
	eventAPI           eventAPI
	scenarioRepository scenarioRepository
	logger             *slog.Logger
	wg                 sync.WaitGroup
	ctx                context.Context
	cancel             context.CancelFunc
}

func NewScenarioService(eventAPI eventAPI, scenarioRepository scenarioRepository, logger *slog.Logger) *ScenarioService {
	ctx, cancel := context.WithCancel(context.Background())
	s := &ScenarioService{
		eventAPI:           eventAPI,
		scenarioRepository: scenarioRepository,
		logger:             logger,
		ctx:                ctx,
		cancel:             cancel,
	}

	return s
}

func (s *ScenarioService) Run(scenarioID string, delaySecond int) error {
	scenario, err := s.scenarioRepository.GetByName(scenarioID)
	if err != nil {
		return fmt.Errorf("error getting scenario %q: %w", scenarioID, err)
	}

	events, err := s.prepareEvents(scenario, s.eventAPI)
	if err != nil {
		return domain.ErrScenarioUnsupportedFormat
	}

	s.wg.Go(func() {
		err := s.playScenario(s.ctx, events, time.Duration(delaySecond)*time.Second)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				s.logger.Info("scenario cancelled", logattr.Error(err))
				return
			}
			s.logger.Error("send scenario failed", logattr.Error(err))
		}
	})

	return nil
}

func (s *ScenarioService) ListNames() ([]string, error) {
	scenarioNames, err := s.scenarioRepository.ListNames()
	if err != nil {
		return nil, fmt.Errorf("scenario list: %w", err)
	}

	return scenarioNames, nil
}

func (s *ScenarioService) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *ScenarioService) prepareEvents(scenario *domain.Scenario, api eventAPI) ([]func(context.Context) error, error) {
	var events []func(context.Context) error
	for _, raw := range scenario.Events {
		var event scenarioEvent

		err := json.Unmarshal(raw, &event)
		if err != nil {
			return nil, fmt.Errorf("error decoding scenario %q: %w", scenario.Name, err)
		}

		switch event.Type {
		case domain.EventTypeLogin:
			var loginEvent oapi.LoginPayload
			err := json.Unmarshal(raw, &loginEvent)
			if err != nil {
				return nil, fmt.Errorf("error decoding login event in scenario %q: %w", scenario.Name, err)
			}
			events = append(events, func(ctx context.Context) error {
				return api.LoginEvent(ctx, event.EventID, loginEvent)
			})
		case domain.EventTypeTransfer:
			var transferEvent oapi.TransferPayload
			err := json.Unmarshal(raw, &transferEvent)
			if err != nil {
				return nil, fmt.Errorf("error decoding transfer event in scenario %q: %w", scenario.Name, err)
			}
			events = append(events, func(ctx context.Context) error {
				return api.TransferEvent(ctx, event.EventID, transferEvent)
			})
		default:
			return nil, fmt.Errorf("unknown event type %q", event.Type)
		}
	}
	return events, nil
}

func (s *ScenarioService) playScenario(ctx context.Context, events []func(context.Context) error, delay time.Duration) error {
	for _, sendEvent := range events {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		err := sendEvent(ctx)
		if err != nil {
			return fmt.Errorf("error sending scenario: %w", err)
		}
	}

	return nil
}
