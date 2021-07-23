package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		"Request with valid secret": {
			RequestURL: fmt.Sprintf("/api/v1/notify?secret=%s", testutils.GetSecret()),
			Method:     "POST",
			SetupAPI: func(api *plugintest.API) *plugintest.API {
				api.On("LogDebug", testutils.GetMockArgumentsWithType("string", 7)...).Return()
				return api
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedHeader:     http.Header{"Content-Type": []string{"application/json"}},
			ExpectedbodyString: `{"status":"OK"}`,
		},
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
