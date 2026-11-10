package domain

import (
	"maps"
	"slices"
)

type FarmRules struct {
	MinClients          int
	MinCollectorSenders int
}

type FarmLogin struct {
	EventID   string
	AccountID string
	ClientID  string
}

type FarmTransfer struct {
	EventID       string
	FromAccountID string
	ToAccountID   string
}

type GroupEvidence struct {
	EventIDs   []string
	AccountIDs []string
}

func (r FarmRules) Detected(logins []FarmLogin) bool {
	clients := make(map[string]struct{}, len(logins))
	for _, login := range logins {
		clients[login.ClientID] = struct{}{}
	}

	return len(clients) >= r.MinClients
}

func (r FarmRules) Evidence(logins []FarmLogin, transfers []FarmTransfer) GroupEvidence {
	events := make([]string, 0, len(logins)+len(transfers))
	members := make(accountSet, len(logins))

	for _, login := range logins {
		events = append(events, login.EventID)
		members[login.AccountID] = struct{}{}
	}

	collectors := r.collectors(members, transfers)

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
