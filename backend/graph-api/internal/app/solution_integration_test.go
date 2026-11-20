//go:build integration

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testGroupID         = "00000000-0000-0000-0000-000000000001"
	solutionPath        = "/solution"
	testResolvedStatus  = "approved"
	testDecisionComment = "preset"
)

type solutionResponse struct {
	Status   string `json:"status"`
	Decision struct {
		Comment   string    `json:"comment"`
		DecidedAt time.Time `json:"decided_at"`
	} `json:"decision"`
}

func TestGroupSolution_Success(t *testing.T) {
	tests := map[string]struct {
		body            string
		expectedStatus  string
		expectedComment string
	}{
		"approved": {
			body:            `{"decision_comment": "Confirmed fraud", "status": "approved"}`,
			expectedStatus:  "approved",
			expectedComment: "Confirmed fraud",
		},
		"declined": {
			body:            `{"decision_comment": "False positive", "status": "declined"}`,
			expectedStatus:  "declined",
			expectedComment: "False positive",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clearEventTables(t)
			insertTestGroup(t)

			before := time.Now().Add(-time.Second)

			resp := patchSolution(t, testGroupID, tt.body)
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var got solutionResponse
			require.NoError(t, json.Unmarshal(body, &got))

			require.Equal(t, tt.expectedStatus, got.Status)
			require.Equal(t, tt.expectedComment, got.Decision.Comment)
			require.False(t, got.Decision.DecidedAt.IsZero(), "decided_at must not be zero")
			require.WithinRange(t, got.Decision.DecidedAt, before, time.Now().Add(time.Second))

			var (
				dbStatus    string
				dbIsFrozen  bool
				dbComment   string
				dbDecidedAt time.Time
			)

			err = testPool.QueryRow(t.Context(),
				`SELECT status, is_frozen, decision_comment, decided_at FROM groups WHERE id = $1`,
				testGroupID,
			).Scan(&dbStatus, &dbIsFrozen, &dbComment, &dbDecidedAt)
			require.NoError(t, err)

			require.Equal(t, tt.expectedStatus, dbStatus)
			require.True(t, dbIsFrozen, "group must be frozen after solution")
			require.Equal(t, tt.expectedComment, dbComment)
			require.Equal(t, got.Decision.DecidedAt.Unix(), dbDecidedAt.Unix(),
				"decided_at in DB must match response")
		})
	}
}

func TestGroupSolution_Error(t *testing.T) {
	tests := map[string]struct {
		resolved       bool
		requestGroupID string
		body           string

		expectedStatus int
	}{
		"group not found": {
			resolved:       false,
			requestGroupID: "00000000-0000-0000-0000-000000000002",
			body:           `{"decision_comment": "Q", "status": "approved"}`,
			expectedStatus: http.StatusNotFound,
		},
		"group already resolved": {
			resolved:       true,
			requestGroupID: testGroupID,
			body:           `{"decision_comment": "Q", "status": "approved"}`,
			expectedStatus: http.StatusConflict,
		},
		"empty comment": {
			resolved:       false,
			requestGroupID: testGroupID,
			body:           `{"decision_comment": "", "status": "approved"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		"wrong status": {
			resolved:       false,
			requestGroupID: testGroupID,
			body:           `{"decision_comment": "Q", "status": "approve"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		"wrong ID": {
			resolved:       false,
			requestGroupID: "group-1",
			body:           `{"decision_comment": "Q", "status": "approved"}`,
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clearEventTables(t)
			insertTestGroup(t)

			if tt.resolved {
				resolveGroupStatus(t)
			}

			resp := patchSolution(t, tt.requestGroupID, tt.body)
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			var (
				dbStatus   string
				dbIsFrozen bool
				dbComment  *string
			)

			err := testPool.QueryRow(t.Context(),
				`SELECT status, is_frozen, decision_comment FROM groups WHERE id = $1`,
				testGroupID,
			).Scan(&dbStatus, &dbIsFrozen, &dbComment)
			require.NoError(t, err)

			if tt.resolved {
				require.Equal(t, testResolvedStatus, dbStatus)
				require.True(t, dbIsFrozen)
				require.NotNil(t, dbComment)
				require.Equal(t, testDecisionComment, *dbComment)
			} else {
				require.Equal(t, "new", dbStatus)
				require.False(t, dbIsFrozen)
				require.Nil(t, dbComment)
			}
		})
	}
}

func patchSolution(t *testing.T, groupID, body string) *http.Response {
	t.Helper()

	path := serverURL + groupsPath + "/" + groupID + solutionPath

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, path, strings.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func insertTestGroup(t *testing.T) {
	t.Helper()

	query := `
		INSERT INTO groups (
			id,
			scheme,
			status,
			risk_score,
			risk_parts,
			first_event_at,
			last_event_at,
			accounts_count,
			total_amount,
		    risk_level
		) VALUES (
			$1,
			'farm',
			'new',
			50,
			'{}'::jsonb,
			'2026-11-19 10:00:00+00',
			'2026-11-19 10:15:00+00',
			1,
			1000,
			'medium'
		)
	`
	_, err := testPool.Exec(t.Context(), query, testGroupID)
	require.NoError(t, err)
}

func resolveGroupStatus(t *testing.T) {
	t.Helper()

	query := `UPDATE groups SET status = $1, is_frozen = TRUE, decided_at = NOW(), decision_comment = $2 WHERE id = $3`

	_, err := testPool.Exec(t.Context(), query, testResolvedStatus, testDecisionComment, testGroupID)
	require.NoError(t, err)
}
