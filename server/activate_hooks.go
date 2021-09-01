package main

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) OnActivate() error {
	if err := p.initBotUser(); err != nil {
		return err
	}

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	p.router = p.InitAPI()

	return nil
}

func (p *Plugin) initBotUser() error {
	botID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    p.configuration.BotUserName,
		DisplayName: p.configuration.BotDisplayName,
		Description: p.configuration.BotDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot")
	}

	p.botID = botID
	return nil
}
