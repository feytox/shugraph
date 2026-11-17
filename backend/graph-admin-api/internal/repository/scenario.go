package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

type ScenarioReadOnlyRepository struct {
	fsys fs.FS
}

func NewScenarioReadOnlyRepository(fsys fs.FS) *ScenarioReadOnlyRepository {
	return &ScenarioReadOnlyRepository{fsys: fsys}
}

func (r *ScenarioReadOnlyRepository) GetByName(name string) (*domain.Scenario, error) {
	fileName := name + ".json"

	data, err := fs.ReadFile(r.fsys, fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, domain.ErrScenarioNotFound
		}

		return nil, fmt.Errorf("read scenario file: %w", err)
	}

	var sc domain.Scenario

	err = json.Unmarshal(data, &sc)
	if err != nil {
		return nil, fmt.Errorf("decode scenario %q: %w", name, err)
	}

	return &sc, nil
}

func (r *ScenarioReadOnlyRepository) ListNames() ([]string, error) {
	entries, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read scenarios dir: %w", err)
	}

	var scenarioNames []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fileName := strings.TrimSuffix(entry.Name(), ".json")
		scenarioNames = append(scenarioNames, fileName)
	}

	return scenarioNames, nil
}
