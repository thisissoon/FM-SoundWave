// React to specific events from a Redis Pub/Sub channel. For example
// Pausing / Playing and Skipping the current Track.
//
// Message bodies will follow this JSON format:
//
// {"event": "EVENT_NAME", "other": "data"}

package soundwave

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/op/go-libspotify/spotify"
	"gopkg.in/redis.v3"
)

// Events we need to listen for
const (
	RESUME_EVENT string = "resume" // Resume paused track
	PAUSE_EVENT  string = "pause"  // Pause a playing track
	STOP_EVENT   string = "stop"   // Stop the currently playing track
)

// Event type to unmarshal message payloads into
type Event struct {
	Type string `json:"event"`
}

// Type for creating an event reactor
type Reactor struct {
	RedisChannelName string
	RedisClient      *redis.Client
	SpotifyPlayer    *spotify.Player
	SpotifySession   *spotify.Session
}

// Subscribes to a redis Pub/Sub channel and consumes messages on the channel
// Once a message is recieved the message it is delegated to the correct
// handler method
func (r *Reactor) Consume() {
	// Subscribe to channel, exiting the program on fail
	pubsub := r.RedisClient.PubSub()
	err := pubsub.Subscribe(r.RedisChannelName)
	if err != nil {
		log.Fatalln(err)
	}

	// Ensure connection the channel is closed on exit
	defer pubsub.Close()

	// Loop to recieve events
	for {
		msg, err := pubsub.Receive() // recieve a message from the channel
		if err != nil {
			log.Println(err)
		} else {
			switch m := msg.(type) {
			case *redis.Subscription:
				log.Println(strings.Title(m.Kind)+":", m.Channel)
			case *redis.Message:
				err := r.processPayload([]byte(m.Payload))
				if err != nil {
					log.Println(err)
				}
			default:
				log.Println("Unknown message: %#v", m)
			}
		}
	}
}

// Processes the raw message payload, unmarshaling the JSON and handing
// the event off to a deligate
func (r *Reactor) processPayload(payload []byte) error {
	// Unmarshal the payload
	e := &Event{}
	err := json.Unmarshal(payload, e)
	if err != nil {
		return err
	}

	// Switch the event type and hand off to handler method
	switch e.Type {
	case PAUSE_EVENT:
		r.pausePlayer()
	case RESUME_EVENT:
		r.resumePlayer()
	case STOP_EVENT:
		r.stopTrack()
	}

	return nil
}

// Pause the Spotify Player
func (r *Reactor) pausePlayer() {
	log.Println("Pause Player")
	// Pause the Track
	r.SpotifyPlayer.Pause()
}

// Resume the Spotify Player
func (r *Reactor) resumePlayer() {
	log.Println("Resume Player")
	// Play the Track
	r.SpotifyPlayer.Play()
}

// Stop the Current Track - Unloading the track causing the next track
// to be played
func (r *Reactor) stopTrack() {
	log.Println("Stop Track")
	// Force the track to stop
	StopTrack <- struct{}{}
}

// Constructor for the Reactor type taking 3 arguments:
// - Redis Channel Name
// - Redis Client
// - Spotify Player
// - Spotify Session
func NewReactor(c string, r *redis.Client, p *spotify.Player) *Reactor {
	return &Reactor{
		RedisChannelName: c,
		RedisClient:      r,
		SpotifyPlayer:    p,
	}
}
