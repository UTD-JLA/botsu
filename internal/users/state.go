package users

import (
	"context"
	"log"
)

// var NotFoundErr = errors.New("user not found")

type UserState struct {
	users map[string]*User
	repo  *UserRepository
}

func NewUserState(r *UserRepository) *UserState {
	return &UserState{repo: r, users: make(map[string]*User)}
}

func (s *UserState) GetUser(id string) (*User, error) {
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
	err := s.repo.Update(context.Background(), user)

	if err != nil {
		return err
	}

	s.users[user.ID] = user
	return nil
}
