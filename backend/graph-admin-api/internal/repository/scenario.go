package repository

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/feytox/shugraph/backend/graph-admin-api/internal/domain"
)

const scenariosDir = "../../graph-admin-api/scenarios"

type ScenarioReadOnlyRepository struct {
}

func NewScenarioReadOnlyRepository() *ScenarioReadOnlyRepository {
	return &ScenarioReadOnlyRepository{}
}

func (*ScenarioReadOnlyRepository) GetByName(name string) (*domain.Scenario, error) {
	fileName := name + ".json"
	data, err := fs.ReadFile(os.DirFS(scenariosDir), fileName)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, domain.ErrScenarioNotFound
		}
		return nil, fmt.Errorf("read scenario file: %w", err)
	}

	var sc domain.Scenario
	errUnmarshal := json.Unmarshal(data, &sc)

	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return &sc, nil
}

func (*ScenarioReadOnlyRepository) ListNames() ([]string, error) {
	files, err := os.ReadDir(scenariosDir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", scenariosDir, err)
	}

	var scenarioNames []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := trimExtension(file.Name())
		scenarioNames = append(scenarioNames, fileName)
	}

	return scenarioNames, nil
}

func trimExtension(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
