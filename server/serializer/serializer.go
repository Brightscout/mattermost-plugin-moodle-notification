package serializer

import (
	"encoding/json"
	"io"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/Brightscout/mattermost-plugin-moodle-notification/server/utils"
)

type Notification struct {
	Email       string `json:"email"`
	Message     string `json:"message"`
	MessageHTML string `json:"messageHTML"`
	Subject     string `json:"subject"`
}

func (notification *Notification) ToJSON() []byte {
	b, _ := json.Marshal(notification)
	return b
}

func (notification *Notification) GetNotificationPost(botID, channelID string) *model.Post {
	converter := md.NewConverter("", true, nil)
	subject := notification.Subject
	message, _ := converter.ConvertString(notification.MessageHTML)

	// This covers the case if notification.MessageHTML is empty
	// or if there is an error while converting the HTML to markdown
	if message == "" {
		message = notification.Message
	}

	if notification.Subject == "" {
		subject = "Moodle Notification"
	}

	post := &model.Post{
		UserId:    botID,
		ChannelId: channelID,
	}

	slackAttachment := &model.SlackAttachment{
		Title: subject,
		Color: "#FF8000",
		Text:  utils.ReplaceEmbeddedImages(message),
	}

	model.ParseSlackAttachment(post, []*model.SlackAttachment{slackAttachment})
	return post
}

func NotificationFromJSON(data io.Reader) (Notification, error) {
	var ps Notification
	err := json.NewDecoder(data).Decode(&ps)
	return ps, err
}
