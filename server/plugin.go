package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
	configurationLock sync.RWMutex
	router            *mux.Router
	botID             string
	configuration     *configuration
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New plugin request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.router.ServeHTTP(w, r)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
