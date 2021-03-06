package handlers

import (
	"app/authz"
	"app/db"
	"app/helpers"
	"app/logs"
	"app/usecase"
	"app/view"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func UserBoards(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
		}
		currentUserID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		boards, err := usecase.UserBoards(data, authLayer, userID, currentUserID)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewBoards(boards))
		if err != nil {
			logs.Error("Request: %s, serializing boards: %v", requestSummary(r), err)
			err := helpers.NewInternalServerError(err)
			ResponseError(w, r, err)
			return
		}

		w.Header().Set(contentType, jsonContent)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}

func ServeUser(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		vars := mux.Vars(r)
		userID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter id: %v", requestSummary(r), err)
			BadRequest(w, r)
			return
		}

		user, err := data.Users().GetUser(userID)
		if err != nil {
			logs.Error("Request: %s, getting user from db: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		currentUserID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			Unauthorized(w, r)
			return
		}

		response := view.NewUser(user)
		if userID != currentUserID {
			response.Email = ""
		}

		bytes, err := json.Marshal(response)
		if err != nil {
			logs.Error("Request: %s, serializing users: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		w.Header().Set(contentType, jsonContent)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}

func UpdateUser(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		vars := mux.Vars(r)
		userID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		currentUserID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		if userID != currentUserID {
			logs.Error("Request: %s, forbidden: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		requestUser := &view.User{}
		if err := json.NewDecoder(r.Body).Decode(requestUser); err != nil {
			logs.Error("Request: %s, unable to parse content: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		user, err := view.NewUserModel(requestUser)
		if err != nil {
			logs.Error("Request: %s, an error occurred: %v", requestSummary(r), err)
		}

		userModel, err := usecase.UpdateUser(data, user)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewUser(userModel))
		if err != nil {
			logs.Error("Request: %s, serializing user: %v", requestSummary(r), err)
			err := helpers.NewInternalServerError(err)
			ResponseError(w, r, err)
			return
		}

		w.Header().Set(contentType, jsonContent)
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}
