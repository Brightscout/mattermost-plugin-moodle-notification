package main

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	BotUserName    = "moodle"
	BotDisplayName = "Moodle"
	BotDescription = "A bot account created by the moodle notification plugin."
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
		Username:    BotUserName,
		DisplayName: BotDisplayName,
		Description: BotDescription,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot")
	}

	p.botID = botID
	return nil
}
