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
	in  chan []byte // channel to read messages from
	out *Channels   // channels to pass events too
}

// Reads messages of the event channel and deligates to other channeks
// to  be actioned upon
func (h *Handler) Run() {
	for {
		msg := <-h.in
		// Unmarshal the message
		e := &event{}
		if err := json.Unmarshal(msg, e); err != nil {
			log.Errorf("Error Unmarshaling %s: %s", msg, err)
		}
		// Switch the event type
		switch e.Type {
		case ADD_EVENT:
			// pass to volume channel
			log.Debugf("Place on Add Channel: %s", msg)
			h.out.Add <- msg
		case PAUSE_EVENT:
			// pass to pause channel
			log.Debugf("Place on Pause Channel: %s", msg)
			h.out.Pause <- msg
		case RESUME_EVENT:
			// pass to resume channel
			log.Debugf("Place on Resume Channel: %s", msg)
			h.out.Resume <- msg
		case STOP_EVENT:
			// pass to stop channel
			log.Debugf("Place on Stop Channel: %s", msg)
			h.out.Stop <- msg
		}
	}
}

func (h *Handler) ReceiveChannel() chan []byte {
	return h.in
}

// Constructs a new Handler
func NewHandler(out *Channels) *Handler {
	return &Handler{
		in:  make(chan []byte),
		out: out,
	}
}
