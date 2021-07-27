package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/Brightscout/mattermost-plugin-moodle-notification/server/serializer"
)

// InitAPI initializes the REST API
func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.Use(p.withRecovery)

	p.handleStaticFiles(r)
	s := r.PathPrefix("/api/v1").Subrouter()

	// Add the custom plugin routes here
	s.HandleFunc("/notify", p.handleNotify).Methods(http.MethodPost)

	// 404 handler
	r.Handle("{anything:.*}", http.NotFoundHandler())
	return r
}

func returnStatusOK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	m := make(map[string]string)
	m[model.STATUS] = model.STATUS_OK
	_, _ = w.Write([]byte(model.MapToJson(m)))
}

func (p *Plugin) handleNotify(w http.ResponseWriter, r *http.Request) {
	if status, err := verifyHTTPSecret(p.configuration.Secret, r.FormValue("secret")); err != nil {
		p.API.LogError(fmt.Sprintf("Invalid Secret. Error: %v", err.Error()))
		http.Error(w, err.Error(), status)
		return
	}

	notification, err := serializer.NotificationFromJSON(r.Body)
	if err != nil {
		p.API.LogError("Error decoding request body.", "Error", err.Error())
		http.Error(w, "Could not decode request body", http.StatusBadRequest)
		return
	}

	if notification.Email == "" || notification.Message == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, appErr := p.API.GetUserByEmail(notification.Email)
	if appErr != nil {
		p.API.LogError(fmt.Sprintf(appErr.Error(), notification.Email))
		http.Error(w, appErr.Error(), http.StatusInternalServerError)
		return
	}

	channel, cErr := p.API.GetDirectChannel(user.Id, p.botID)
	if cErr != nil {
		p.API.LogError(fmt.Sprintf(cErr.Error(), notification.Email))
		http.Error(w, cErr.Error(), http.StatusInternalServerError)
		return
	}

	_, pErr := p.API.CreatePost(&model.Post{
		UserId:    p.botID,
		ChannelId: channel.Id,
		Message:   notification.Message,
	})

	if pErr != nil {
		p.API.LogError(fmt.Sprintf("Could not send DM to user: %v", pErr.Error()))
		http.Error(w, fmt.Sprintf("Could not send DM to user: %v", pErr.Error()), http.StatusInternalServerError)
		return
	}

	returnStatusOK(w)
}

// handleStaticFiles handles the static files under the assets directory.
func (p *Plugin) handleStaticFiles(r *mux.Router) {
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		p.API.LogWarn("Failed to get bundle path.", "Error", err.Error())
		return
	}

	// This will serve static files from the 'assets' directory under '/static/<filename>'
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(bundlePath, "assets")))))
}

// withRecovery allows recovery from panics
func (p *Plugin) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				p.API.LogError("Recovered from a panic",
					"url", r.URL.String(),
					"error", x,
					"stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Ref: mattermost plugin confluence(https://github.com/mattermost/mattermost-plugin-confluence/blob/3ee2aa149b6807d14fe05772794c04448a17e8be/server/controller/main.go#L97)
func verifyHTTPSecret(expected, got string) (status int, err error) {
	for {
		if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1 {
			break
		}

		unescaped, _ := url.QueryUnescape(got)
		if unescaped == got {
			return http.StatusForbidden, errors.New("request URL: secret did not match")
		}
		got = unescaped
	}

	return 0, nil
}
