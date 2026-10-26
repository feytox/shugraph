package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type GroupHandler struct{}

func (h *GroupHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "list-groups", Summary: "List suspicious groups", Method: http.MethodGet,
		Path: "/api/v1/groups", Errors: []int{400, 422, 501},
		Description: "Ordered by first_event_at DESC, id DESC. cursor is an opaque continuation token bound to this order. Poll the first page to discover new groups. Invalid cursors return 400. An empty list has items: [] and next_cursor: null.",
	}, h.List)
	huma.Register(api, huma.Operation{
		OperationID: "get-group", Summary: "Get a suspicious group", Method: http.MethodGet,
		Path: "/api/v1/groups/{group_id}", Errors: []int{404, 422, 501},
	}, h.GetByID)
	huma.Register(api, huma.Operation{
		OperationID: "get-group-graph", Summary: "Get a group's Cytoscape elements", Method: http.MethodGet,
		Path: "/api/v1/groups/{group_id}/graph", Errors: []int{404, 422, 501},
		Description: "Only the selected group's evidence subgraph. IDs are prefixed by entity/edge type. Parallel transfers remain separate edges. Each login produces two edges sharing event_id: account to device and account to network address. Ownership is client to account. Layout and presentation are client concerns.",
	}, h.GetGraph)
	huma.Register(api, huma.Operation{
		OperationID: "resolve-group", Summary: "Confirm or decline a new group", Method: http.MethodPatch,
		Path: "/api/v1/groups/{group_id}/solution", Errors: []int{400, 404, 409, 422, 501},
		Description: "approved=true confirms, false declines. A non-whitespace comment is required. Both decisions freeze the group and set decided_at. Only status новая may be decided, including a group frozen by expiry. Any subsequent decision returns 409; undo/history are outside this contract.",
	}, h.ApplySolution)
}

type GroupSchemeType string

const (
	GroupSchemeFarm    GroupSchemeType = "ферма"
	GroupSchemeAddress GroupSchemeType = "адрес"
	GroupSchemeChain   GroupSchemeType = "цепочка"
	GroupSchemeRing    GroupSchemeType = "кольцо"
)

type GroupStatusType string

const (
	GroupStatusNew      GroupStatusType = "новая"
	GroupStatusApproved GroupStatusType = "подтверждена"
	GroupStatusDeclined GroupStatusType = "отклонена"
)

type Feature struct {
	AccountsCount int   `json:"accounts_count" minimum:"0"`
	SpanMinutes   int   `json:"span_minutes" minimum:"0" doc:"Elapsed whole minutes between first and last evidence events."`
	TotalAmount   int64 `json:"total_amount" minimum:"0" doc:"Turnover in kopecks: sum of all evidence transfers, including repeated rounds."`
}

type GroupItem struct {
	ID           string          `json:"id" format:"uuid"`
	Scheme       GroupSchemeType `json:"scheme" enum:"ферма,адрес,цепочка,кольцо"`
	Status       GroupStatusType `json:"status" enum:"новая,подтверждена,отклонена"`
	RiskScore    int             `json:"risk_score" minimum:"0" maximum:"100"`
	RiskParts    map[string]int  `json:"risk_parts" doc:"Scheme-specific contributions; their sum equals risk_score. Used to explain the finding."`
	FirstEventAt time.Time       `json:"first_event_at"`
	LastEventAt  time.Time       `json:"last_event_at"`
	Features     Feature         `json:"features"`
}

type Group struct {
	GroupItem
	IsFrozen        bool       `json:"is_frozen"`
	DecisionComment string     `json:"decision_comment" doc:"Empty until a decision is made."`
	DecidedAt       *time.Time `json:"decided_at" nullable:"true" doc:"Null while status is новая."`
	Members         []Account  `json:"members" nullable:"false"`
	Events          []Event    `json:"events" nullable:"false"`
}

type GroupListOutput struct {
	Body struct {
		Items      []GroupItem `json:"items" nullable:"false"`
		NextCursor *string     `json:"next_cursor" nullable:"true" doc:"Pass unchanged as cursor; null on the last page."`
	}
}

type GroupIDInput struct {
	GroupID string `path:"group_id" format:"uuid"`
}

type ListGroupInput struct {
	Limit  int    `query:"limit" default:"10" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor" doc:"Opaque continuation token. Omit for the first page."`
}

type GetByIDOutput struct {
	Body struct {
		Group Group `json:"group"`
	}
}

type GetGraphOutput struct{ Body GroupGraph }

type GroupSolutionInput struct {
	GroupID string `path:"group_id" format:"uuid"`
	Body    struct {
		Approved        bool   `json:"approved"`
		DecisionComment string `json:"decision_comment" minLength:"1" pattern:"\\S"`
	}
}

type GroupSolutionOutput struct {
	Body struct {
		Status          GroupStatusType `json:"status" enum:"подтверждена,отклонена"`
		IsFrozen        bool            `json:"is_frozen" enum:"true"`
		DecisionComment string          `json:"decision_comment" minLength:"1" pattern:"\\S"`
		DecidedAt       time.Time       `json:"decided_at"`
	}
}

func (h *GroupHandler) List(context.Context, *ListGroupInput) (*GroupListOutput, error) {
	return nil, huma.Error501NotImplemented("Group listing is not implemented yet.")
}
func (h *GroupHandler) GetByID(context.Context, *GroupIDInput) (*GetByIDOutput, error) {
	return nil, huma.Error501NotImplemented("Group retrieval is not implemented yet.")
}
func (h *GroupHandler) GetGraph(context.Context, *GroupIDInput) (*GetGraphOutput, error) {
	return nil, huma.Error501NotImplemented("Group graph retrieval is not implemented yet.")
}
func (h *GroupHandler) ApplySolution(context.Context, *GroupSolutionInput) (*GroupSolutionOutput, error) {
	return nil, huma.Error501NotImplemented("Group decisions are not implemented yet.")
}
