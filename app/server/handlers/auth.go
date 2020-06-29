package handlers

import (
	"app/assets"
	"app/authz"
	"app/db"
	"app/helpers"
	"app/logs"
	"app/models"
	"app/view"
	"encoding/json"
	"net/http"
)

func SignIn(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		signInRequest, err := view.NewSignInRequest(r.Body)
		if err != nil {
			logs.Error("Request: %s, getting user data from req body: %v", requestSummary(r), err)
			BadRequest(w, r)
		}

		token, err := authLayer.AuthenticateUser(signInRequest.Email, signInRequest.Password)
		if err != nil {
			logs.Error("Request: %s, authenticate user: %v", requestSummary(r), err)
			Unauthorized(w, r)
			return
		}
		userID, err := authz.GetUserIDByToken(authLayer, token)

		response := view.NewLSignInResponse(token, userID)
		bytes, err := json.Marshal(response)
		if err != nil {
			logs.Error("Request: %s, serializing login response: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		w.Header().Set(contentType, jsonContent)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}

func SignUp(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		signUpRequest, err := view.NewSignUpRequest(r.Body)
		if err != nil {
			logs.Error("Request: %s, getting user data from req body: %v", requestSummary(r), err)
			BadRequest(w, r)
			return
		}

		hashedPassword, err := helpers.HashPassword(signUpRequest.Password)
		if err != nil {
			logs.Error("Request: %s, hashing password: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		user := &models.User{
			Name:           signUpRequest.Name,
			Email:          signUpRequest.Email,
			HashedPassword: hashedPassword,
			Icon:           assets.UserIcon, // TODO: replace with image url on S3
		}
		_, err = data.Users().CreateUser(user)
		if err != nil {
			logs.Error("Request: %s, creating user: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		token, err := authLayer.AuthenticateUser(signUpRequest.Email, signUpRequest.Password)
		if err != nil {
			logs.Error("Request: %s, authenticate user: %v", requestSummary(r), err)
			Unauthorized(w, r)
			return
		}
		userID, err := authz.GetUserIDByToken(authLayer, token)

		response := view.NewLSignUpResponse(token, userID)
		bytes, err := json.Marshal(response)
		if err != nil {
			logs.Error("Request: %s, serializing sign up response: %v", requestSummary(r), err)
			InternalServerError(w, r)
			return
		}

		w.Header().Set(contentType, jsonContent)
		w.WriteHeader(http.StatusCreated)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}
