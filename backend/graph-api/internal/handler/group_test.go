package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDecisionResponseShape(t *testing.T) {
	decision := Decision{
		Comment:   "Confirmed",
		DecidedAt: time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC),
	}
	solution := GroupSolutionOutput{}
	solution.Body.Status = GroupStatusApproved
	solution.Body.Decision = decision

	for _, tt := range []struct {
		name string
		body any
		want string
	}{
		{"new group", Group{}, `null`},
		{"decided group", Group{Decision: &decision}, `{"comment":"Confirmed","decided_at":"2026-10-29T12:00:00Z"}`},
		{"solution", solution.Body, `{"comment":"Confirmed","decided_at":"2026-10-29T12:00:00Z"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.body)
			require.NoError(t, err)
			var fields map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(data, &fields))
			require.Contains(t, fields, "decision")
			require.JSONEq(t, tt.want, string(fields["decision"]))
			require.NotContains(t, fields, "decision_comment")
			require.NotContains(t, fields, "decided_at")
		})
	}
}
