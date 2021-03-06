package usecase

import (
	"app/db"
	"app/helpers"
	"app/logs"
	"app/models"
	"fmt"
)

func CreateBoard(data db.DataStorageInterface, board *models.Board) (*models.Board, helpers.AppError) {

	board, err := data.Boards().CreateBoard(board)
	if err != nil {
		logs.Error("An error occurred while creating board: %v", err)
		return nil, helpers.NewInternalServerError(err)
	}

	return board, nil
}

func UpdateBoard(data db.DataStorageInterface, board *models.Board) (*models.Board, error) {

	b, err := data.Boards().GetBoard(board.ID)
	if err != nil {
		logs.Error("An error occurred: %v", err)
		return nil, helpers.NewBadRequest(err)
	}

	if b.UserID != board.UserID {
		err := fmt.Errorf("UserIDs do not match error")
		logs.Error("An error occurred: %v", err)
		return nil, helpers.NewUnauthorized(err)
	}

	if err := data.Boards().UpdateBoard(board); err != nil {
		logs.Error("An error occurred: %v", err)
		return nil, helpers.NewInternalServerError(err)
	}

	return board, nil
}

func SavePin(data db.DataStorageInterface, boardID int, pinID int) helpers.AppError {
	// Check board and pin exist
	_, err := data.Boards().GetBoard(boardID)
	if err != nil {
		logs.Error("An error occurred while checking board exists: %v", err)
		return helpers.NewBadRequest(err)
	}

	_, err = data.Pins().GetPin(pinID)
	if err != nil {
		logs.Error("An error occurred while checking pin exists: %v", err)
		return helpers.NewBadRequest(err)
	}

	// TODO: Check if board-pin row already exists
	// See: https://github.com/team-e-org/backend/issues/242

	err = data.BoardsPins().CreateBoardPin(boardID, pinID)
	if err != nil {
		logs.Error("An error occurred while adding board pin: %v", err)
		return helpers.NewInternalServerError(err)
	}

	return nil
}

func UnsavePin(data db.DataStorageInterface, boardID int, pinID int) helpers.AppError {
	err := data.BoardsPins().DeleteBoardPin(boardID, pinID)
	if err != nil {
		logs.Error("An error occurred while deleting board pin: %v", err)
		return helpers.NewInternalServerError(err)
	}

	return nil
}
