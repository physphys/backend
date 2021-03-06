package handlers

import (
	"app/authz"
	"app/db"
	"app/helpers"
	"app/logs"
	"app/models"
	"app/ptr"
	"app/repository"
	"app/usecase"
	"app/view"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func ServePinsInBoard(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		userID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		vars := mux.Vars(r)
		boardID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		page, err := strconv.Atoi(r.FormValue("page"))
		if err != nil {
			logs.Error("Request: %s, parse path parameter page: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		pins, err := usecase.GetPinsByBoardID(data, userID, boardID, page)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewPins(pins))
		if err != nil {
			logs.Error("Request: %s, serializing pins: %v", requestSummary(r), err)
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

func ServePinsByTag(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		tag := r.FormValue("tag")
		if tag == "" {
			err := fmt.Errorf("No tag given")
			logs.Error("Request: %s, parse path parameter tag: %v", requestSummary(r), err)
			err = helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		var page int
		var err error
		page, err = strconv.Atoi(r.FormValue("page"))
		if err != nil {
			page = 1
		}

		pins, err := usecase.GetPinsByTag(data, tag, page)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			err := helpers.NewInternalServerError(err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewPins(pins))
		if err != nil {
			logs.Error("Request: %s, serializing pins: %v", requestSummary(r), err)
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

func ServePins(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		var page int
		var err error
		page, err = strconv.Atoi(r.FormValue("page"))
		if err != nil {
			page = 1
		}

		pins, err := usecase.GetPins(data, page)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewPins(pins))
		if err != nil {
			logs.Error("Request: %s, serializing pins: %v", requestSummary(r), err)
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

func ServeHomePins(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		userID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		homePinsRequest := &view.HomePinsRequest{}
		json.NewDecoder(r.Body).Decode(homePinsRequest)

		pins, nextPagingKey, err := usecase.GetHomePins(data, userID, homePinsRequest.PagingKey)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		res := view.HomePinsResponse{
			Pins:      view.NewPins(pins),
			PagingKey: nextPagingKey,
		}
		bytes, err := json.Marshal(res)
		if err != nil {
			logs.Error("Request: %s, serializing pins: %v", requestSummary(r), err)
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

func ServePin(data db.DataStorageInterface, authLayer authz.AuthLayerInterface) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		userID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		vars := mux.Vars(r)
		pinID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		fmt.Printf("pinID: %d, userID: %d\n", pinID, userID)
		pin, err := usecase.ServePin(data, pinID, userID)
		fmt.Printf("pin: %+v\n", pin)
		if err != nil {
			logs.Error("Request: %s, %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		bytes, err := json.Marshal(view.NewPin(pin))

		if err != nil {
			logs.Error("Request: %s, serializing pin: %v", requestSummary(r), err)
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

func CreatePin(data db.DataStorageInterface, authLayer authz.AuthLayerInterface, lambda repository.LambdaRepository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		userID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		maxSize := int64(1024000)
		err = r.ParseMultipartForm(maxSize)
		if err != nil {
			logs.Error("Request: %s, parsing multipart: %v", requestSummary(r), err)
			logs.Error("Image too large. Max Size: %v", maxSize)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		vars := mux.Vars(r)
		boardID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter board id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		var b bool
		if r.FormValue("isPrivate") != "" {
			b, err = strconv.ParseBool(r.FormValue("isPrivate"))
			if err != nil {
				logs.Error("Request: %s, parse parameter isPrivate: %v", requestSummary(r), err)
				err := helpers.NewBadRequest(err)
				ResponseError(w, r, err)
				return
			}
		}

		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			logs.Error("Request: %s, getting uploaded image file: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}
		defer file.Close()

		fileExt := filepath.Ext(fileHeader.Filename)
		fileName := data.AWSS3().CreateFileName(userID, fileExt)

		pin := &models.Pin{
			UserID:      ptr.NewInt(userID),
			Title:       r.FormValue("title"),
			Description: ptr.NewString(r.FormValue("description")),
			URL:         ptr.NewString(r.FormValue("url")),
			IsPrivate:   b,
			ImageURL:    fileName,
		}

		pin, err = usecase.CreatePin(data, pin, file, fileName, fileExt, boardID)
		if err != nil {
			logs.Error("Request: %s, an error occurred: %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		baseURL := data.AWSS3().GetBaseURL()
		imageURL := fmt.Sprintf("%s/%s", baseURL, pin.ImageURL)
		pin.ImageURL = imageURL

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

		go func() {
			var tags []string
			if r.FormValue("tags") != "" {
				tags = strings.Split(r.FormValue("tags"), " ")
			}

			if len(tags) > 0 {
				logs.Info("attaching tags, %+v to %+v", tags, pin)
				err = lambda.AttachTagsWithContext(ctx, pin, tags)
				if err != nil {
					logs.Error("Request: %s, invoke attachTags lambda failed : %v", requestSummary(r), err)
				}
				cancel()
			}
		}()

		viewPin := view.NewPin(pin)
		bytes, err := json.Marshal(viewPin)
		if err != nil {
			logs.Error("Request: %s, serializing pin response: %v", requestSummary(r), err)
			err := helpers.NewInternalServerError(err)
			ResponseError(w, r, err)
			return
		}

		w.Header().Set(contentType, jsonContent)
		w.WriteHeader(http.StatusCreated)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}

func UpdatePin(data db.DataStorageInterface, authLayer authz.AuthLayerInterface, lambda repository.LambdaRepository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		userID, err := getUserIDIfAvailable(r, authLayer)
		if err != nil {
			logs.Error("Request: %s, checking if user identifiable: %v", requestSummary(r), err)
			err := helpers.NewUnauthorized(err)
			ResponseError(w, r, err)
			return
		}

		vars := mux.Vars(r)
		pinID, err := strconv.Atoi(vars["id"])
		if err != nil {
			logs.Error("Request: %s, parse path parameter board id: %v", requestSummary(r), err)
			err := helpers.NewBadRequest(err)
			ResponseError(w, r, err)
			return
		}

		pin := &models.Pin{
			ID:          pinID,
			Title:       r.FormValue("title"),
			Description: ptr.NewString(r.FormValue("description")),
			URL:         ptr.NewString(r.FormValue("url")),
		}

		pin, err = usecase.UpdatePin(data, pin, userID)
		if err != nil {
			logs.Error("Request: %s, an error occurred: %v", requestSummary(r), err)
			ResponseError(w, r, err)
			return
		}

		baseURL := data.AWSS3().GetBaseURL()
		imageURL := fmt.Sprintf("%s/%s", baseURL, pin.ImageURL)
		pin.ImageURL = imageURL

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

		go func() {
			var tags []string
			if r.FormValue("tags") != "" {
				tags = strings.Split(r.FormValue("tags"), " ")
			}

			if len(tags) > 0 {
				logs.Info("attaching tags, %+v to %+v", tags, pin)
				err = lambda.AttachTagsWithContext(ctx, pin, tags)
				if err != nil {
					logs.Error("Request: %s, invoke attachTags lambda failed : %v", requestSummary(r), err)
				}
				cancel()
			}
		}()

		response := view.NewPin(pin)
		bytes, err := json.Marshal(response)
		if err != nil {
			logs.Error("Request: %s, serializing pin response: %v", requestSummary(r), err)
			err := helpers.NewInternalServerError(err)
			ResponseError(w, r, err)
			return
		}

		w.Header().Set(contentType, jsonContent)
		w.WriteHeader(http.StatusCreated)
		if _, err = w.Write(bytes); err != nil {
			logs.Error("Request: %s, writing response: %v", requestSummary(r), err)
		}
	}
}
