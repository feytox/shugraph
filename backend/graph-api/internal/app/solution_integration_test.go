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
	testGroupID  = "00000000-0000-0000-0000-000000000001"
	groupPath    = "/api/v1/groups/"
	solutionPath = "/solution"
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
			clearGroups(t)
			insertTestGroup(t, testGroupID)

			before := time.Now()

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
		existedGroupID    string
		groupStatusBefore string
		requestGroupID    string
		body              string

		expectedStatus int
	}{
		"group not found": {
			testGroupID,
			"new",
			"00000000-0000-0000-0000-000000000002",
			`{"decision_comment": "Q", "status": "approved"}`,
			http.StatusNotFound,
		},
		"group already resolved": {
			testGroupID,
			"approved",
			testGroupID,
			`{"decision_comment": "Q", "status": "approved"}`,
			http.StatusConflict,
		},
		"empty comment": {
			testGroupID,
			"new",
			testGroupID,
			`{"decision_comment": "", "status": "approved"}`,
			http.StatusUnprocessableEntity,
		},
		"wrong status": {
			testGroupID,
			"new",
			testGroupID,
			`{"decision_comment": "Q", "status": "approve"}`,
			http.StatusUnprocessableEntity,
		},
		"wrong ID": {
			testGroupID,
			"new",
			"group-1",
			`{"decision_comment": "Q", "status": "approved"}`,
			http.StatusUnprocessableEntity,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clearGroups(t)
			insertTestGroup(t, tt.existedGroupID)

			if tt.groupStatusBefore != "new" {
				updateGroupStatus(t, tt.existedGroupID, tt.groupStatusBefore)
			}

			resp := patchSolution(t, tt.requestGroupID, tt.body)
			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func patchSolution(t *testing.T, groupID, body string) *http.Response {
	t.Helper()

	path := serverURL + groupPath + groupID + solutionPath

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, path, strings.NewReader(body))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func clearGroups(t *testing.T) {
	t.Helper()

	_, err := testPool.Exec(t.Context(), `TRUNCATE groups CASCADE`)
	require.NoError(t, err)
}

func insertTestGroup(t *testing.T, groupID string) {
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
			total_amount
		) VALUES (
			$1,
			'farm',
			'new',
			50,
			'{}'::jsonb,
			'2026-11-19 10:00:00+00',
			'2026-11-19 10:15:00+00',
			1,
			1000
		)
	`
	_, err := testPool.Exec(t.Context(), query, groupID)
	require.NoError(t, err)
}

func updateGroupStatus(t *testing.T, groupID string, status string) {
	t.Helper()

	query := `UPDATE groups SET status = $1, decided_at = NOW(), decision_comment = 'preset' WHERE id = $2`

	_, err := testPool.Exec(t.Context(), query, status, groupID)
	require.NoError(t, err)
}
