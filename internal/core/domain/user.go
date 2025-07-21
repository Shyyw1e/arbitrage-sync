package domain

import (
	"fmt"
	"sync"
)

type UserState struct {
	MinDiff float64
	MaxSum   float64
	Step    string		//"waiting_foe_input", "ready_to_run", etc.
}


type UserStatesStore struct {
	mu sync.RWMutex
	store map[int64]*UserState
}

func NewUserStatesStore() *UserStatesStore {
	return &UserStatesStore{
		store: make(map[int64]*UserState),
	}
}

func (s *UserStatesStore) Get(chatID int64) (*UserState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.store[chatID]
	if !ok {
		err := fmt.Errorf("no user state with such chat ID: %v", ok)
		return val, err
	}
	return val, nil	
}

func (s *UserStatesStore) Set(chatID int64, state *UserState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[chatID] = state
}

func (s *UserStatesStore) Delete(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, chatID)
}