package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/Brightscout/mattermost-plugin-moodle-notification/server/serializer"
	"github.com/Brightscout/mattermost-plugin-moodle-notification/server/utils/testutils"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
)

func setupTestPlugin(api *plugintest.API) *Plugin { //nolint:interfacer
	p := &Plugin{}
	p.setConfiguration(&configuration{
		Secret: testutils.GetSecret(),
	})

	path, _ := filepath.Abs("../..")
	api.On("GetBundlePath").Return(path, nil)

	p.SetAPI(api)
	p.router = p.InitAPI()

	return p
}

func TestServeHTTP(t *testing.T) {
	for name, test := range map[string]struct {
		RequestURL         string
		Method             string
		SetupAPI           func(*plugintest.API) *plugintest.API
		ExpectedStatusCode int
		ExpectedHeader     http.Header
		ExpectedbodyString string
	}{
		"Request with invalid secret": {
			RequestURL: fmt.Sprintf("/api/v1/notify?secret=%s", "1232323rsdsdf"),
			Method:     "POST",
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("LogDebug", testutils.GetMockArgumentsWithType("string", 7)...).Return()
				api.On("LogError", testutils.GetMockArgumentsWithType("string", 1)...).Return()
				return api
			},
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedHeader:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}, "X-Content-Type-Options": []string{"nosniff"}},
			ExpectedbodyString: "request URL: secret did not match\n",
		},
		"InvalidRequestURL": {
			RequestURL: "/not_found",
			Method:     "GET",
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("LogDebug", testutils.GetMockArgumentsWithType("string", 7)...).Return()
				return api
			},
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedHeader:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}, "X-Content-Type-Options": []string{"nosniff"}},
			ExpectedbodyString: "404 page not found\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			api := test.SetupAPI(&plugintest.API{})
			defer api.AssertExpectations(t)
			p := setupTestPlugin(api)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(test.Method, test.RequestURL, nil)
			p.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)
			defer result.Body.Close()

			bodyBytes, err := ioutil.ReadAll(result.Body)
			require.Nil(t, err)
			bodyString := string(bodyBytes)

			assert.Equal(test.ExpectedbodyString, bodyString)
			assert.Equal(test.ExpectedStatusCode, result.StatusCode)
			assert.Equal(test.ExpectedHeader, result.Header)
		})
	}
}

func TestHandleNotify(t *testing.T) {
	email := "test@example.com"
	expectedUser := &model.User{Email: email, Username: "test"}

	for name, test := range map[string]struct {
		RequestURL         string
		SetupAPI           func(*plugintest.API) *plugintest.API
		Request            *serializer.Notification
		ExpectedStatusCode int
		ExpectedHeader     http.Header
		ExpectedbodyString string
	}{
		"Request with invalid email": {
			RequestURL: fmt.Sprintf("/api/v1/notify?secret=%s", testutils.GetSecret()),
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("LogDebug", testutils.GetMockArgumentsWithType("string", 7)...).Return()
				api.On("LogError", testutils.GetMockArgumentsWithType("string", 1)...).Return()
				// return error on get user by email.
				api.On("GetUserByEmail", email).Return(nil, &model.AppError{
					Where:         "GetUserByEmail",
					Message:       "Unable to find the user.",
					DetailedError: "resource: User id: email=test@example.com",
				})
				return api
			},
			Request: &serializer.Notification{
				Message: "Notification from moodle",
				Email:   expectedUser.Email,
			},
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedHeader:     http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}, "X-Content-Type-Options": []string{"nosniff"}},
			ExpectedbodyString: "GetUserByEmail: Unable to find the user., resource: User id: email=test@example.com\n",
		},
		"Request with valid email": {
			RequestURL: fmt.Sprintf("/api/v1/notify?secret=%s", testutils.GetSecret()),
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("LogDebug", testutils.GetMockArgumentsWithType("string", 7)...).Return()
				api.On("GetUserByEmail", email).Return(expectedUser, nil)
				api.On("GetDirectChannel", testutils.GetMockArgumentsWithType("string", 2)...).Return(&model.Channel{Id: "channelID"}, nil)
				api.On("CreatePost", testutils.GetMockArgumentsWithType("*model.Post", 1)...).Return(&model.Post{}, nil)
				return api
			},
			Request: &serializer.Notification{
				Message: "Notification from moodle",
				Email:   expectedUser.Email,
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"status":"OK"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			api := test.SetupAPI(&plugintest.API{})
			defer api.AssertExpectations(t)
			p := setupTestPlugin(api)

			body := bytes.NewReader(test.Request.ToJSON())

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", test.RequestURL, body)
			p.ServeHTTP(nil, w, r)

			result := w.Result()
			require.NotNil(t, result)
			defer result.Body.Close()

			bodyBytes, err := ioutil.ReadAll(result.Body)
			require.Nil(t, err)
			bodyString := string(bodyBytes)

			assert.Equal(test.ExpectedbodyString, bodyString)
			assert.Equal(test.ExpectedStatusCode, result.StatusCode)
			assert.Equal(test.ExpectedHeader, result.Header)
		})
	}
}
