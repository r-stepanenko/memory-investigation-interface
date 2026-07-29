package search

import "event-memory-search-api/internal/domain"

type UserIndex struct {
	ByUser map[string][]domain.Event
}

func BuildUserIndex(events []domain.Event) *UserIndex {
	index := &UserIndex{
		ByUser: make(map[string][]domain.Event),
	}

	for _, event := range events {
		index.ByUser[event.UserID] = append(
			index.ByUser[event.UserID],
			event,
		)
	}

	return index
}

func (u *UserIndex) Find(user string) []domain.Event {
	if user == "" {
		return nil
	}

	return u.ByUser[user]
}