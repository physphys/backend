package server

import (
	"app/authz"
	"app/db"
	"app/goldenfiles"
	"app/models"
	"app/ptr"
	"app/testutils/dbdata"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	helpers "app/testutils"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestServePin(t *testing.T) {
	var cases = []struct {
		Desc  string
		Code  int
		pinID int
		pin   *models.Pin
	}{
		{
			"single pin",
			200,
			1,
			pin1(),
		},
		{
			"no pin",
			404, // testでは500になる。sql.ErrNoRowsのため
			2,
			pin1(),
		},
	}

	for _, c := range cases {
		t.Run(helpers.TableTestName(c.Desc), func(t *testing.T) {
			router := mux.NewRouter()
			data := db.NewRepositoryMock()

			data.Pins().CreatePin(c.pin, 0)

			attachHandlers(router, data, authz.NewAuthLayerMock(data))
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/pins/%d", c.pinID), nil)

			router.ServeHTTP(recorder, req)
			body := recorder.Body.Bytes()

			assert.Equal(t, c.Code, recorder.Code, "Status code should match reference")
			expected := goldenfiles.UpdateAndOrRead(t, body)
			assert.Equal(t, expected, body, "Response body should match golden file")
		})
	}
}

func TestServePinsInBoard(t *testing.T) {
	var cases = []struct {
		Desc    string
		Code    int
		boardID int
		pins    []*models.Pin
		page    int
	}{
		{
			"single pin",
			200,
			1,
			[]*models.Pin{pin1()},
			1,
		},
		{
			"two pin",
			200,
			1,
			[]*models.Pin{pin1(), pin2()},
			1,
		},
		{
			"private pin",
			200,
			1,
			[]*models.Pin{pin1(), pin2(), privatePin()},
			1,
		},
		{
			"no pin",
			200,
			1,
			[]*models.Pin{privatePin()},
			1,
		},
	}

	for _, c := range cases {
		t.Run(helpers.TableTestName(c.Desc), func(t *testing.T) {
			router := mux.NewRouter()
			data := db.NewRepositoryMock()

			for _, p := range c.pins {
				data.Pins().CreatePin(p, c.boardID)
			}

			attachHandlers(router, data, authz.NewAuthLayerMock(data))
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/boards/%d/pins?page=%d", c.boardID, c.page), nil)

			router.ServeHTTP(recorder, req)
			body := recorder.Body.Bytes()

			assert.Equal(t, c.Code, recorder.Code, "Status code should match reference")
			expected := goldenfiles.UpdateAndOrRead(t, body)
			assert.Equal(t, expected, body, "Response body should match golden file")
		})
	}
}

func TestCreatePin(t *testing.T) {
	var cases = []struct {
		Desc          string
		Code          int
		filePath      string
		fieldValues   map[string]string
		currentUser   *models.User
		loginPassword string
	}{
		{
			"success",
			201,
			"./testdata/sample.png",
			map[string]string{
				"title":       "test title",
				"description": "test description",
			},
			dbdata.BaseUser,
			dbdata.BaseUserPassword,
		},
	}

	for _, c := range cases {
		t.Run(helpers.TableTestName(c.Desc), func(t *testing.T) {
			router := mux.NewRouter()
			data := db.NewRepositoryMock()

			data.Users().CreateUser(c.currentUser)

			al := authz.NewAuthLayerMock(data)
			token, _ := al.AuthenticateUser(c.currentUser.Email, c.loginPassword)

			attachReqAuth(router, data, al)
			recorder := httptest.NewRecorder()

			requestBody, contentType := buildMulitpartRequest(t, c.filePath, c.fieldValues)
			req := httptest.NewRequest(http.MethodPost, "/boards/0/pins", requestBody)
			req.Header.Set("Content-Type", contentType)
			helpers.SetAuthTokenHeader(req, token)

			router.ServeHTTP(recorder, req)
			body := recorder.Body.Bytes()

			assert.Equal(t, c.Code, recorder.Code, "Status code should match reference")
			expected := goldenfiles.UpdateAndOrRead(t, body)
			assert.Equal(t, expected, body, "Response body should match golden file")
		})
	}
}

func pin1() *models.Pin {
	return &models.Pin{
		ID:          1,
		UserID:      ptr.NewInt(0),
		Title:       "test title",
		Description: ptr.NewString("test description"),
		URL:         ptr.NewString("test url"),
		ImageURL:    "test image url",
		IsPrivate:   false,
	}
}

func pin2() *models.Pin {
	return &models.Pin{
		ID:          2,
		UserID:      ptr.NewInt(0),
		Title:       "test title 2",
		Description: ptr.NewString("test description2"),
		URL:         ptr.NewString("test url2"),
		ImageURL:    "test image url2",
		IsPrivate:   false,
	}
}

func privatePin() *models.Pin {
	return &models.Pin{
		ID:          3,
		UserID:      ptr.NewInt(1),
		Title:       "test title  private",
		Description: ptr.NewString("test description private"),
		URL:         ptr.NewString("test url private"),
		ImageURL:    "test image url private",
		IsPrivate:   true,
	}
}

func buildMulitpartRequest(t *testing.T, imageFilePath string, fieldValues map[string]string) (*bytes.Buffer, string) {
	file, err := os.Open(imageFilePath)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	defer mw.Close()

	fw, err := mw.CreateFormFile("image", "sample")
	if err != nil {
		t.Error(err)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Error(err)
	}

	for k, v := range fieldValues {
		mw.WriteField(k, v)
	}

	return body, mw.FormDataContentType()
}
