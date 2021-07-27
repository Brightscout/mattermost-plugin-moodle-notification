package serializer

import (
	"encoding/json"
	"io"
)

type Notification struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

func (n *Notification) ToJSON() []byte {
	b, _ := json.Marshal(n)
	return b
}

func NotificationFromJSON(data io.Reader) (Notification, error) {
	var ps Notification
	err := json.NewDecoder(data).Decode(&ps)
	return ps, err
}
