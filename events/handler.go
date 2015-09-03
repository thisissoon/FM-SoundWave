// Event Handler

package events

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
)

type event struct {
	Type string `json:"event"`
}

type Handler struct {
	channel chan []byte // channel to read messages from
}

// Reads messages of the event channel and deligates to other channeks
// to  be actioned upon
func (h *Handler) Run() {
	for {
		msg := <-h.channel
		// Unmarshal the message
		e := &envent{}
		if err := json.Unmarshal(msg, e); err != nil {
			log.Errorf("Error Unmarshaling Message: %s", err)
		}
		// Switch the event type
		switch e.Type {
		case ADD_EVENT:
			// pass to volume channel
		case PAUSE_EVENT:
			// pass to pause channel
		case RESUME_EVENT:
			// pass to resume channel
		case STOP_EVENT:
			// pass to stop channel
		}
	}
}

// Constructs a new Handler
func NewHandler(c chan []byte) *EventHandler {
	return &EventHandler{
		channel: c,
	}
}
