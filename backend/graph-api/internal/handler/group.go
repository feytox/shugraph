package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/feytox/shugraph/backend/graph-api/internal/domain"
	"github.com/feytox/shugraph/backend/graph-api/internal/service"
)

const groupsTag = "Groups"

type GroupService interface {
	List(ctx context.Context, filter domain.GroupFilter) (service.GroupPage, error)
	Get(ctx context.Context, id uuid.UUID) (service.GroupDetails, error)
}

type groupHandlers struct {
	groups GroupService
	log    *slog.Logger
}

func registerGroups(api huma.API, h *groupHandlers) {
	huma.Register(api, huma.Operation{
		OperationID: "list-groups",
		Summary:     "List suspicious groups",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusBadRequest, http.StatusUnprocessableEntity},
	}, h.list)

	huma.Register(api, huma.Operation{
		OperationID: "get-group",
		Summary:     "Get a suspicious group",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_id}",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusNotFound, http.StatusUnprocessableEntity},
	}, h.getByID)

	huma.Register(api, huma.Operation{
		OperationID: "get-group-graph",
		Summary:     "Get a group's evidence graph",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_id}/graph",
		Tags:        []string{groupsTag},
		Errors:      []int{http.StatusNotFound, http.StatusUnprocessableEntity},
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
			http.StatusNotImplemented,
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
	GroupID uuid.UUID `path:"group_id" format:"uuid"`
}

type ListGroupInput struct {
	Limit  int    `query:"limit" default:"10" minimum:"1" maximum:"100"`
	Cursor string `query:"cursor" doc:"Opaque continuation token. Omit for the first page."`
	Status string `query:"status" doc:"Status to filter groups to." enum:"new,approved,declined"`

	after *domain.GroupCursor
}

func (in *ListGroupInput) Resolve(huma.Context) []error {
	if in.Cursor == "" {
		return nil
	}

	cursor, err := decodeCursor(in.Cursor)
	if err != nil {
		return []error{&huma.ErrorDetail{Location: "query.cursor", Message: "invalid cursor", Value: in.Cursor}}
	}

	in.after = &cursor

	return nil
}

func (in *ListGroupInput) filter() domain.GroupFilter {
	filter := domain.GroupFilter{After: in.after, Limit: in.Limit}

	if in.Status != "" {
		status := domain.GroupStatus(in.Status)
		filter.Status = &status
	}

	return filter
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

func (h *groupHandlers) list(ctx context.Context, in *ListGroupInput) (*GroupListOutput, error) {
	page, err := h.groups.List(ctx, in.filter())
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	out := &GroupListOutput{}
	out.Body.Items = make([]GroupItem, 0, len(page.Groups))

	for _, group := range page.Groups {
		item, err := groupItem(group)
		if err != nil {
			return nil, problem(ctx, h.log, err)
		}

		out.Body.Items = append(out.Body.Items, item)
	}

	if page.Next != nil {
		next := encodeCursor(*page.Next)
		out.Body.NextCursor = &next
	}

	return out, nil
}

func (h *groupHandlers) getByID(ctx context.Context, in *GroupIDInput) (*GetByIDOutput, error) {
	details, err := h.groups.Get(ctx, in.GroupID)
	if err != nil {
		return nil, h.groupProblem(ctx, err)
	}

	group, err := groupCard(details)
	if err != nil {
		return nil, problem(ctx, h.log, err)
	}

	return &GetByIDOutput{Body: group}, nil
}

func (h *groupHandlers) getGraph(ctx context.Context, in *GroupIDInput) (*GetGraphOutput, error) {
	details, err := h.groups.Get(ctx, in.GroupID)
	if err != nil {
		return nil, h.groupProblem(ctx, err)
	}

	return &GetGraphOutput{Body: groupGraph(details)}, nil
}

func (h *groupHandlers) applySolution(context.Context, *GroupSolutionInput) (*GroupSolutionOutput, error) {
	return nil, huma.Error501NotImplemented("Group decisions are not implemented yet.")
}

func (h *groupHandlers) groupProblem(ctx context.Context, err error) error {
	if errors.Is(err, domain.ErrGroupNotFound) {
		return huma.Error404NotFound(domain.ErrGroupNotFound.Error())
	}

	return problem(ctx, h.log, err)
}

func groupItem(group domain.Group) (GroupItem, error) {
	parts, err := riskParts(group.Scheme, group.RiskParts)
	if err != nil {
		return GroupItem{}, fmt.Errorf("group %s: %w", group.ID, err)
	}

	return GroupItem{
		ID:           group.ID.String(),
		Scheme:       GroupSchemeType(group.Scheme),
		Status:       GroupStatusType(group.Status),
		RiskLevel:    GroupRiskType(group.RiskLevel),
		RiskScore:    group.RiskScore,
		RiskParts:    parts,
		DetectedAt:   group.DetectedAt,
		FirstEventAt: group.FirstEventAt,
		LastEventAt:  group.LastEventAt,
		Features: Feature{
			AccountsCount: group.AccountsCount,
			SpanMinutes:   group.SpanMinutes(),
			TotalAmount:   group.TotalAmount,
		},
	}, nil
}

func riskParts(scheme domain.GroupScheme, raw json.RawMessage) (RiskParts, error) {
	var (
		parts  RiskParts
		target any
	)

	switch scheme {
	case domain.GroupSchemeFarm:
		parts.Farm = &FarmScheme{}
		target = parts.Farm
	case domain.GroupSchemeAddress:
		parts.Address = &AddressScheme{}
		target = parts.Address
	case domain.GroupSchemeChain:
		parts.Chain = &ChainScheme{}
		target = parts.Chain
	case domain.GroupSchemeRing:
		parts.Ring = &RingScheme{}
		target = parts.Ring
	default:
		return RiskParts{}, fmt.Errorf("unknown scheme %q", scheme)
	}

	if err := json.Unmarshal(raw, target); err != nil {
		return RiskParts{}, fmt.Errorf("decode risk parts: %w", err)
	}

	return parts, nil
}

func groupCard(details service.GroupDetails) (Group, error) {
	item, err := groupItem(details.Group)
	if err != nil {
		return Group{}, err
	}

	group := Group{
		GroupItem: item,
		IsFrozen:  details.Group.IsFrozen,
		Members:   make([]Account, 0, len(details.Members)),
		Logins:    make([]LoginEvent, 0, len(details.Logins)),
		Transfers: make([]TransferEvent, 0, len(details.Transfers)),
	}

	if decision := details.Group.Decision; decision != nil {
		group.Decision = &Decision{Comment: decision.Comment, DecidedAt: decision.DecidedAt}
	}

	for _, member := range details.Members {
		group.Members = append(group.Members, Account{ID: member.ID, ClientID: member.ClientID})
	}

	for _, login := range details.Logins {
		group.Logins = append(group.Logins, LoginEvent{
			EventID:    login.EventID,
			OccurredAt: login.OccurredAt,
			AccountID:  login.AccountID,
			ClientID:   login.ClientID,
			DeviceID:   login.DeviceID,
			AddressID:  login.AddressID,
		})
	}

	for _, transfer := range details.Transfers {
		group.Transfers = append(group.Transfers, TransferEvent{
			EventID:       transfer.EventID,
			OccurredAt:    transfer.OccurredAt,
			FromAccountID: transfer.FromAccountID,
			FromClientID:  transfer.FromClientID,
			ToAccountID:   transfer.ToAccountID,
			ToClientID:    transfer.ToClientID,
			Amount:        transfer.Amount,
		})
	}

	return group, nil
}
