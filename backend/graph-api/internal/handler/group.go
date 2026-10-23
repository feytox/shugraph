package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type GroupHandler struct {
}

func (h *GroupHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "list-groups",
		Summary:       "Returns list of groups",
		Method:        http.MethodGet,
		Path:          "/api/v1/groups",
		DefaultStatus: http.StatusOK,
	}, h.List)

	huma.Register(api, huma.Operation{
		OperationID:   "get-group",
		Summary:       "Return a group",
		Method:        http.MethodGet,
		Path:          "/api/v1/groups/{group_id}",
		DefaultStatus: http.StatusOK,
	}, h.GetByID)

	huma.Register(api, huma.Operation{
		OperationID:   "get-group-graph",
		Summary:       "Returns the vertices and edges of the group graph",
		Method:        http.MethodGet,
		Path:          "/api/v1/groups/{group_id}/graph",
		DefaultStatus: http.StatusOK,
	}, h.GetGraph)

	huma.Register(api, huma.Operation{
		OperationID:   "resolve-group",
		Summary:       "Makes decisions for the group",
		Method:        http.MethodPatch,
		Path:          "/api/v1/groups/{group_id}/solution",
		DefaultStatus: http.StatusOK,
	}, h.ApplySolution)
}

type GroupSchemeType string

const (
	GroupSchemeFarm    GroupSchemeType = "Ферма"
	GroupSchemeAddress GroupSchemeType = "Адрес"
	GroupSchemeChain   GroupSchemeType = "Цепочка"
	GroupSchemeRing    GroupSchemeType = "Кольцо"
)

type GroupItem struct {
	ID        string
	Scheme    GroupSchemeType
	RiskScore int
}

type GroupListOutput struct {
	Body struct {
		Items []GroupItem
	}
}
type GroupIDInput struct {
	GroupID uuid.UUID `path:"group_id"`
}

type ListGroupInput struct {
	GroupID uuid.UUID `path:"group_id"`
	Limit   int       `query:"limit" default:"10" maximum:"100"`
	Cursor  int       `query:"cursor" default:"0"`
}

func (h *GroupHandler) List(_ context.Context, _ *ListGroupInput) (*GroupListOutput, error) {
	return nil, nil
}

type Feature struct {
	AccountsCount int
	SpanMinutes   int
	TotalAmount   int
}

type Group struct {
	ID              uuid.UUID
	Scheme          GroupSchemeType
	Status          GroupStatusType
	RiskScore       int
	RiskParts       map[string]int
	FirstEventAt    time.Time
	LastEventAt     time.Time
	IsFrozen        bool
	DecisionComment string
	Features        []Feature
}

type GetByIDOutput struct {
	Body struct {
		group Group
	}
}

func (h *GroupHandler) GetByID(_ context.Context, _ *GroupIDInput) (*GetByIDOutput, error) {
	return nil, nil
}

type GroupGraph struct {
	// TODO
}

type GetGraphOutput struct {
	Body struct {
		// TODO
	}
}

func (h *GroupHandler) GetGraph(_ context.Context, _ *GroupIDInput) (*GetGraphOutput, error) {
	return nil, nil
}

type GroupSolutionInput struct {
	GroupID string `path:"group_id"`
	Body    struct {
		Approved        *bool
		DecisionComment *string
	}
}
type GroupStatusType string

const (
	GroupStatusNew      GroupStatusType = "Новая"
	GroupStatusApproved GroupStatusType = "Подтверждена"
	GroupStatusDeclined GroupStatusType = "Отклонена"
)

type GroupSolutionOutput struct {
	Body struct {
		Status          GroupStatusType
		IsFrozen        bool
		DecisionComment string
	}
}

func (h *GroupHandler) ApplySolution(_ context.Context, _ *GroupSolutionInput) (*GroupSolutionOutput, error) {
	_ = GroupStatusNew
	_ = GroupStatusApproved
	_ = GroupStatusDeclined
	_ = GroupSchemeFarm
	_ = GroupSchemeAddress
	_ = GroupSchemeChain
	_ = GroupSchemeRing
	// ub for fast commit raviles caude code fix if needed
	return nil, nil
}
