package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/google/uuid"
)

const groupsTag = "Groups"

type GroupSolutionService interface {
	ApplySolution(ctx context.Context, solution domain.GroupSolution) (domain.GroupSolution, error)
}

type groupHandlers struct {
	solutionService GroupSolutionService
	log             *slog.Logger
}

func registerGroups(api huma.API, h *groupHandlers) {
	huma.Register(api, huma.Operation{
		OperationID: "list-groups",
		Summary:     "List suspicious groups",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusNotImplemented},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-group",
		Summary:     "Get a suspicious group",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_id}",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusNotImplemented},
	}, h.getByID)

	huma.Register(api, huma.Operation{
		OperationID: "get-group-graph",
		Summary:     "Get a group's evidence graph",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_id}/graph",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusNotImplemented},
	}, h.getGraph)

	huma.Register(api, huma.Operation{
		OperationID: "resolve-group",
		Summary:     "Confirm or decline a new group",
		Method:      http.MethodPatch,
		Path:        "/api/v1/groups/{group_id}/solution",
		Tags:        []string{groupsTag},
		Errors: []int{
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusConflict,
			http.StatusUnprocessableEntity,
		},
	}, h.applySolution)
}

type GroupRiskType string

const (
	GroupRiskLow    GroupRiskType = "low"
	GroupRiskMedium GroupRiskType = "medium"
	GroupRiskHigh   GroupRiskType = "high"
)

type GroupSchemeType string

const (
	GroupSchemeFarm    GroupSchemeType = "farm"
	GroupSchemeAddress GroupSchemeType = "address"
	GroupSchemeChain   GroupSchemeType = "chain"
	GroupSchemeRing    GroupSchemeType = "ring"
)

type GroupStatusType string

const (
	GroupStatusNew      GroupStatusType = "new"
	GroupStatusApproved GroupStatusType = "approved"
	GroupStatusDeclined GroupStatusType = "declined"
)

type Feature struct {
	AccountsCount int   `json:"accounts_count" minimum:"0"`
	SpanMinutes   int   `json:"span_minutes" minimum:"0" doc:"Elapsed whole minutes between first and last evidence events."`
	TotalAmount   int64 `json:"total_amount" minimum:"0" doc:"Turnover in kopecks: sum of all evidence transfers, including repeated rounds."`
}

type RiskParts struct {
	Farm    *FarmScheme    `json:"farm,omitempty" doc:"'Farm' scheme metrics"`
	Address *AddressScheme `json:"address,omitempty" doc:"'Address' Scheme Metrics"`
	Chain   *ChainScheme   `json:"chain,omitempty" doc:"'Chain' scheme metrics"`
	Ring    *RingScheme    `json:"ring,omitempty" doc:"'Ring' scheme metrics"`
}

type FarmScheme struct {
	Scheme    int `json:"scheme"`
	Speed     int `json:"speed"`
	Clients   int `json:"clients"`
	Collector int `json:"collector"`
}

type AddressScheme struct {
	Scheme  int `json:"scheme"`
	Speed   int `json:"speed"`
	Clients int `json:"clients"`
	Devices int `json:"devices"`
	Burst   int `json:"burst"`
}

type ChainScheme struct {
	Scheme    int `json:"scheme"`
	Speed     int `json:"speed"`
	Hops      int `json:"hops"`
	Preserved int `json:"preserved"`
}

type RingScheme struct {
	Scheme    int `json:"scheme"`
	Speed     int `json:"speed"`
	Rounds    int `json:"rounds"`
	Accounts  int `json:"accounts"`
	Preserved int `json:"preserved"`
}

type GroupItem struct {
	ID           string          `json:"id" format:"uuid"`
	Scheme       GroupSchemeType `json:"scheme" enum:"farm,address,chain,ring"`
	Status       GroupStatusType `json:"status" enum:"new,approved,declined"`
	RiskLevel    GroupRiskType   `json:"risk_level" enum:"low,medium,high"`
	RiskScore    int             `json:"risk_score" minimum:"0" maximum:"100"`
	RiskParts    RiskParts       `json:"risk_parts"`
	DetectedAt   time.Time       `json:"detected_at" doc:"Server time of first detection; immutable."`
	FirstEventAt time.Time       `json:"first_event_at"`
	LastEventAt  time.Time       `json:"last_event_at"`
	Features     Feature         `json:"features"`
}

type Decision struct {
	Comment   string    `json:"comment"`
	DecidedAt time.Time `json:"decided_at"`
}
type Group struct {
	GroupItem
	IsFrozen  bool            `json:"is_frozen"`
	Decision  *Decision       `json:"decision,omitempty"`
	Members   []Account       `json:"members" nullable:"false"`
	Logins    []LoginEvent    `json:"logins" nullable:"false"`
	Transfers []TransferEvent `json:"transfers" nullable:"false"`
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
	Status string `query:"status" doc:"Status to filter groups to." enum:"new,approved,declined"`
}

type GetByIDOutput struct {
	Body Group
}

type GetGraphOutput struct{ Body GroupGraph }

type GroupSolutionInput struct {
	GroupID string `path:"group_id" format:"uuid"`
	Body    struct {
		Status          GroupStatusType `json:"status" enum:"approved,declined"`
		DecisionComment string          `json:"decision_comment" minLength:"1" pattern:"\\S"`
	}
}

type GroupSolutionOutput struct {
	Body struct {
		Status   GroupStatusType `json:"status" enum:"approved,declined"`
		Decision Decision        `json:"decision"`
	}
}

func (h *groupHandlers) list(context.Context, *ListGroupInput) (*GroupListOutput, error) {
	return nil, huma.Error501NotImplemented("Group listing is not implemented yet.")
}
func (h *groupHandlers) getByID(context.Context, *GroupIDInput) (*GetByIDOutput, error) {
	return nil, huma.Error501NotImplemented("Group retrieval is not implemented yet.")
}
func (h *groupHandlers) getGraph(context.Context, *GroupIDInput) (*GetGraphOutput, error) {
	return nil, huma.Error501NotImplemented("Group graph retrieval is not implemented yet.")
}

func (h *groupHandlers) applySolution(ctx context.Context, i *GroupSolutionInput) (*GroupSolutionOutput, error) {
	solution := domain.GroupSolution{
		GroupID:  uuid.MustParse(i.GroupID),
		Decision: &domain.Decision{Comment: i.Body.DecisionComment},
		Status:   domain.GroupStatus(i.Body.Status),
	}

	appliedSolution, err := h.solutionService.ApplySolution(ctx, solution)
	if err != nil {
		if e, ok := errors.AsType[domain.ErrGroupNotFound](err); ok {
			return nil, huma.Error404NotFound(e.Error())
		}

		if e, ok := errors.AsType[domain.ErrGroupSolutionAlreadyExists](err); ok {
			return nil, huma.Error409Conflict(e.Error())
		}

		return nil, problem(ctx, h.log, err)
	}

	out := &GroupSolutionOutput{}

	out.Body.Status = GroupStatusType(appliedSolution.Status)
	out.Body.Decision = Decision{
		Comment:   appliedSolution.Decision.Comment,
		DecidedAt: appliedSolution.Decision.DecidedAt,
	}

	return out, nil
}
