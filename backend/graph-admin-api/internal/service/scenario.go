package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-admin-api/internal/graphapi/oapi"
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
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	Amount     int64     `json:"amount"`
}

type ScenarioService struct {
	eventAPI           eventAPI
	scenarioRepository scenarioRepository
	logger             *slog.Logger
	wg                 sync.WaitGroup
}

func NewScenarioService(eventAPI eventAPI, scenarioRepository scenarioRepository, logger *slog.Logger) *ScenarioService {
	s := &ScenarioService{
		eventAPI:           eventAPI,
		scenarioRepository: scenarioRepository,
		logger:             logger,
	}
	s.wg = sync.WaitGroup{}
	return s
}

func (s *ScenarioService) Run(ctx context.Context, scenarioID string, delaySecond int) error {
	scenario, err := s.scenarioRepository.GetByName(scenarioID)
	if err != nil {
		return fmt.Errorf("error getting scenario %q: %w", scenarioID, err)
	}

	go func() {
		s.wg.Add(1)
		err := s.sendScenario(ctx, scenarioID, scenario, time.Duration(delaySecond)*time.Second)
		if err != nil {
			s.logger.Error("error sending scenario: %w", err)
		}
	}()

	return nil
}

func (s *ScenarioService) ListNames() ([]string, error) {
	scenarioNames, err := s.scenarioRepository.ListNames()
	if err != nil {
		return nil, fmt.Errorf("scenario list: %w", err)
	}
	return scenarioNames, nil
}

func (s *ScenarioService) sendScenario(ctx context.Context, scenarioID string, scenario *domain.Scenario, delay time.Duration) error {
	defer s.wg.Done()
	for _, raw := range scenario.Events {
		time.Sleep(delay)
		event, err := decodeScenarioEvent(raw)
		if err != nil {
			return fmt.Errorf("error decoding scenario %q: %w", scenarioID, err)
		}
		switch event.Type {
		case domain.EventTypeLogin:
			if err := handleEvent(ctx, event.EventID, raw, s.eventAPI.LoginEvent); err != nil {
				return fmt.Errorf("send login event %q for scenario %q: %w", event.Type, scenarioID, err)
			}
		case domain.EventTypeTransfer:
			if err := handleEvent(ctx, event.EventID, raw, s.eventAPI.TransferEvent); err != nil {
				return fmt.Errorf("send transfer event %q for scenario %q: %w", event.Type, scenarioID, err)
			}
		default:
			return fmt.Errorf("unknown event type %q", event.Type)
		}
	}
	return nil
}

func handleEvent[T any](ctx context.Context, eventID string, raw json.RawMessage,
	send func(context.Context, string, T) error) error {
	var event T
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	return send(ctx, eventID, event)
}

func decodeScenarioEvent(raw json.RawMessage) (scenarioEvent, error) {
	var event scenarioEvent
	err := json.Unmarshal(raw, &event)
	if err != nil {
		return scenarioEvent{}, err
	}

	return event, nil
}
