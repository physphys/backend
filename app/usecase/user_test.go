package usecase

import (
	"app/authz"
	"app/db"
	"app/helpers"
	"app/models"
	"app/ptr"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestUserBoard(t *testing.T) {
	data := db.NewRepositoryMock()
	boards := []*models.Board{
		{
			ID:          0,
			UserID:      0,
			Name:        "test board",
			Description: ptr.NewString("test description"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          1,
			UserID:      1,
			Name:        "test board2",
			Description: ptr.NewString("test description2"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			UserID:      0,
			Name:        "test board3",
			Description: ptr.NewString("test description3"),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	for _, b := range boards {
		_, err := data.Boards().CreateBoard(b)
		if err != nil {
			t.Fatalf("An error occurred: %v", err)
		}
	}
	authz := authz.NewAuthLayerMock(data)
	userID := 0
	currentUserID := 0
	boards, err := UserBoards(data, authz, userID, currentUserID)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}
	for _, b := range boards {
		if b.UserID != userID {
			t.Fatalf("Could not get user's boards")
		}
	}
}

func TestEmptyUserBoardError(t *testing.T) {
	data := db.NewRepositoryMock()
	authz := authz.NewAuthLayerMock(data)
	userID := 0
	currentUserID := 0
	_, err := UserBoards(data, authz, userID, currentUserID)
	if err == nil {
		t.Fatalf("Board is empty")
	}
}

func TestRemovePrivateBoards(t *testing.T) {
	boards := []*models.Board{
		{
			ID:          0,
			UserID:      0,
			Name:        "test board",
			Description: ptr.NewString("test description"),
			IsPrivate:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          1,
			UserID:      0,
			Name:        "test board2",
			Description: ptr.NewString("test description2"),
			IsPrivate:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			UserID:      1,
			Name:        "test board3",
			Description: ptr.NewString("test description3"),
			IsPrivate:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          3,
			UserID:      1,
			Name:        "test board4",
			Description: ptr.NewString("test description4"),
			IsPrivate:   true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	userID := 0
	boards = removePrivateBoards(boards, userID)

	if len(boards) != 3 {
		t.Fatalf("len(boards) should be 3")
	}

	for _, b := range boards {
		if b.UserID != userID && b.IsPrivate {
			t.Fatalf("Other people's private boards are gotten.")
		}
	}
}

func TestUpdateUser(t *testing.T) {
	data := db.NewRepositoryMock()
	password := "password"
	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}
	now := time.Now()
	user := &models.User{
		ID:             1,
		Name:           "test user",
		Email:          "test@test.com",
		Icon:           "test icon",
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	err = data.Users().CreateUser(user)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}

	user2 := &models.User{
		ID:             1,
		Name:           "test user2",
		Email:          "test2@test.com",
		Icon:           "test icon2",
		HashedPassword: hashedPassword,
	}
	user3, err := UpdateUser(data, user2)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}
	user, err = data.Users().GetUser(1)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}
	if !reflect.DeepEqual(user, user3) {
		t.Fatalf("User is not updated")
	}
}

func TestUpdateUserError(t *testing.T) {
	password := "password"
	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		t.Fatalf("An error occurred: %v", err)
	}

	now := time.Now()
	data := db.NewRepositoryMock()
	user := &models.User{
		ID:             1,
		Name:           "test user",
		Email:          "test@test.com",
		Icon:           "test icon",
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	_, err = UpdateUser(data, user)
	if err == nil {
		t.Fatalf("An error should occur")
	}

	data.Users().CreateUser(user)
	user2 := &models.User{
		ID:             2,
		Name:           "test user",
		Email:          "test@test.com",
		Icon:           "test icon",
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	data.Users().CreateUser(user2)
	fmt.Printf("user2: %v\n", *user2)
	_, err = UpdateUser(data, user2)
	if err == nil {
		t.Fatalf("An error should occur")
	}
}
