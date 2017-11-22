package log

import "encoding/json"

// Message use to format json for Channel.Message
type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Mode    string `json:"mode"`
}

// NewMessage creates a new instance of Message
func NewMessage(id string, c string, isRead bool) *Message {
	m := &Message{
		ID:      id,
		Content: c,
		Mode:    "w",
	}

	if isRead {
		m.Mode = "r"
	}

	return m
}

// JSON converts *MessageLog to json
func (m *Message) JSON() []byte {
	b, _ := json.Marshal(m)
	return b
}
