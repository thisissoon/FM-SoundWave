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

	"gopkg.in/redis.v3"
)

// Events we need to listen for
const (
	ADD_EVENT    string = "add"    // Add track event
	RESUME_EVENT string = "resume" // Resume paused track
	PAUSE_EVENT  string = "pause"  // Pause a playing track
	STOP_EVENT   string = "stop"   // Stop the currently playing track
)

const PAUSE_STATE_KEY = "fm:player:paused"

// Event type to unmarshal message payloads into
type Event struct {
	Type string `json:"event"`
}

// Add event JSON data structure
type AddEvent struct {
	Type string `json:"event"`
	Uri  string `json:"uri"`
	User string `json:"user"`
}

// Type for creating an event reactor
type Reactor struct {
	RedisChannelName string
	RedisClient      *redis.Client
	Player           *Player
}

// Constructor for the Reactor type taking 3 arguments:
// - Redis Channel Name
// - Pointer to Redis Client
// - Pointer to Player  TODO: Remove and switch to channel notification
func NewReactor(c string, r *redis.Client, p *Player) *Reactor {
	return &Reactor{
		RedisChannelName: c,
		RedisClient:      r,
		Player:           p,
	}
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
			switch m := msg.(type) { // Switch the mesage type
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
	case ADD_EVENT:
		return r.addEventHandler(payload)
	case PAUSE_EVENT:
		return r.pauseEventHandler()
	case RESUME_EVENT:
		return r.resumeEventHandler()
	case STOP_EVENT:
		return r.stopEventHandler()
	}

	return nil
}

// Handles add events. This will trigger the player to start prefetching the
// track into the cache
func (r *Reactor) addEventHandler(payload []byte) error {
	i := &AddEvent{}
	err := json.Unmarshal(payload, i)
	if err != nil {
		return err
	}

	return r.Player.Prefetch(&i.Uri)
}

// Pause event handler. This will trigger the player to pause
func (r *Reactor) pauseEventHandler() error {
	log.Println("Pause Event")
	// Pause the Track
	r.Player.Pause() // TODO: Use Channel
	// Set the Redis Key for Storing Player Pause State
	err := r.RedisClient.Set(PAUSE_STATE_KEY, "1", 0).Err()

	return err
}

// Resume event handler will trigger the player to resume playing
// the track
func (r *Reactor) resumeEventHandler() error {
	log.Println("Resume Event")
	// Play the Track
	r.Player.Resume() // TODO: Use Channel
	// Set the Redis Key for Storing Player Pause State
	err := r.RedisClient.Set(PAUSE_STATE_KEY, "0", 0).Err()

	return err
}

// Stop event handler will stop the current track forcing the player to
// play the next track
func (r *Reactor) stopEventHandler() error {
	log.Println("Stop Event")

	// Force the track to stop by placing a message on the StopTrack
	// channel which will cause the Player method to unblock
	StopTrack <- struct{}{}

	return nil // always return nil since no errors can happen here
}
