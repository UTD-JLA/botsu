package users

import (
	"context"
	"log"
	"sync"
)

type UserState struct {
	usersMu sync.RWMutex
	users   map[string]*User
	repo    *UserRepository
}

func NewUserState(r *UserRepository) *UserState {
	return &UserState{repo: r, users: make(map[string]*User)}
}

func (s *UserState) GetUser(id string) (*User, error) {
	s.usersMu.RLock()
	defer s.usersMu.RUnlock()

	user, ok := s.users[id]

	if ok {
		log.Printf("Found user %s in cache", id)
		return user, nil
	}

	user, err := s.repo.FindOrCreate(context.Background(), id)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserState) UpdateUser(user *User) error {
	s.usersMu.Lock()
	defer s.usersMu.Unlock()

	err := s.repo.Update(context.Background(), user)

	if err != nil {
		return err
	}

	s.users[user.ID] = user
	return nil
}
