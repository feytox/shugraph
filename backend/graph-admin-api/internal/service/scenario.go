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

type preparedEvent struct {
	id   string
	send func(context.Context) error
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

func (s *ScenarioService) Run(scenarioID string, delay time.Duration) error {
	scenario, err := s.scenarioRepository.GetByName(scenarioID)
	if err != nil {
		return fmt.Errorf("error getting scenario %q: %w", scenarioID, err)
	}

	events, err := s.prepareEvents(scenario)
	if err != nil {
		return fmt.Errorf("prepare scenario %q: %w", scenarioID, err)
	}

	s.wg.Go(func() {
		err := s.playScenario(s.ctx, events, delay)
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

func (s *ScenarioService) prepareEvents(scenario *domain.Scenario) ([]preparedEvent, error) {
	var events []preparedEvent

	for _, raw := range scenario.Events {
		var event scenarioEvent

		err := json.Unmarshal(raw, &event)
		if err != nil {
			return nil, fmt.Errorf("decoding scenario %q: %w", scenario.Name, err)
		}

		switch event.Type {
		case domain.EventTypeLogin:
			var loginEvent oapi.LoginPayload

			err := json.Unmarshal(raw, &loginEvent)
			if err != nil {
				return nil, fmt.Errorf("decoding login event %q: %w", event.EventID, err)
			}

			send := func(ctx context.Context) error {
				return s.eventAPI.LoginEvent(ctx, event.EventID, loginEvent)
			}
			events = append(events, preparedEvent{event.EventID, send})
		case domain.EventTypeTransfer:
			var transferEvent oapi.TransferPayload

			err := json.Unmarshal(raw, &transferEvent)
			if err != nil {
				return nil, fmt.Errorf("decoding transfer event %q: %w", event.EventID, err)
			}

			send := func(ctx context.Context) error {
				return s.eventAPI.TransferEvent(ctx, event.EventID, transferEvent)
			}
			events = append(events, preparedEvent{event.EventID, send})
		default:
			return nil, fmt.Errorf("unknown event type %q", event.Type)
		}
	}

	return events, nil
}

func (s *ScenarioService) playScenario(ctx context.Context, events []preparedEvent, delay time.Duration) error {
	for _, event := range events {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		err := event.send(ctx)
		if err != nil {
			return fmt.Errorf("send event %q: %w", event.id, err)
		}
	}

	return nil
}
