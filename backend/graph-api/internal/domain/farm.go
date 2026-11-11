package domain

import (
	"cmp"
	"encoding/json"
	"maps"
	"slices"
	"time"
)

type FarmRules struct {
	MinClients          int
	MinCollectorSenders int
}

type FarmLogin struct {
	EventID    string
	OccurredAt time.Time
	AccountID  string
	ClientID   string
}

type FarmTransfer struct {
	EventID       string
	FromAccountID string
	ToAccountID   string
}

type Farm struct {
	Window TimeWindow
	Logins []FarmLogin
}

type GroupEvidence struct {
	EventIDs   []string
	AccountIDs []string
}

func (r GroupRules) DetectFarm(logins []FarmLogin, at time.Time) (Farm, bool) {
	sorted := slices.SortedFunc(slices.Values(logins), func(a, b FarmLogin) int {
		return cmp.Or(a.OccurredAt.Compare(b.OccurredAt), cmp.Compare(a.EventID, b.EventID))
	})

	clients := make(map[string]int)
	end := 0

	for start, first := range sorted {
		window := TimeWindow{Start: first.OccurredAt, End: first.OccurredAt.Add(r.Window)}

		for ; end < len(sorted) && window.Contains(sorted[end].OccurredAt); end++ {
			clients[sorted[end].ClientID]++
		}

		if window.Contains(at) && len(clients) >= r.Farm.MinClients {
			return Farm{Window: window, Logins: sorted[start:end]}, true
		}

		clients[first.ClientID]--
		if clients[first.ClientID] == 0 {
			delete(clients, first.ClientID)
		}
	}

	return Farm{}, false
}

func (f Farm) AccountIDs() []string {
	return slices.Sorted(maps.Keys(f.members()))
}

func (r GroupRules) FarmEvidence(farm Farm, transfers []FarmTransfer) GroupEvidence {
	events := make([]string, 0, len(farm.Logins)+len(transfers))
	for _, login := range farm.Logins {
		events = append(events, login.EventID)
	}

	members := farm.members()
	collectors := r.Farm.collectors(members, transfers)

	for _, transfer := range transfers {
		if members.has(transfer.FromAccountID) && collectors.has(transfer.ToAccountID) {
			events = append(events, transfer.EventID)
		}
	}

	accounts := maps.Clone(members)
	maps.Copy(accounts, collectors)
	slices.Sort(events)

	return GroupEvidence{EventIDs: events, AccountIDs: slices.Sorted(maps.Keys(accounts))}
}

func FarmRisk(GroupStats) Risk {
	return Risk{Score: 0, Parts: json.RawMessage(`{}`)}
}

func (f Farm) members() accountSet {
	members := make(accountSet, len(f.Logins))
	for _, login := range f.Logins {
		members[login.AccountID] = struct{}{}
	}

	return members
}

func (r FarmRules) collectors(members accountSet, transfers []FarmTransfer) accountSet {
	senders := make(map[string]accountSet)

	for _, transfer := range transfers {
		if !members.has(transfer.FromAccountID) {
			continue
		}

		if senders[transfer.ToAccountID] == nil {
			senders[transfer.ToAccountID] = make(accountSet)
		}

		senders[transfer.ToAccountID][transfer.FromAccountID] = struct{}{}
	}

	collectors := make(accountSet)

	for account, from := range senders {
		if len(from) >= r.MinCollectorSenders {
			collectors[account] = struct{}{}
		}
	}

	return collectors
}

type accountSet map[string]struct{}

func (s accountSet) has(account string) bool {
	_, ok := s[account]
	return ok
}
