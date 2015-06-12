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

// Channel for add track eveents
var AddTrack chan struct{}

// Events we need to listen for
const (
	ADD_EVENT    string = "add"    // Add track to queue event
	RESUME_EVENT string = "resume" // Resume paused track
	PAUSE_EVENT  string = "pause"  // Pause a playing track
	STOP_EVENT   string = "stop"   // Stop the currently playing track
)

const PAUSE_STATE_KEY = "fm:player:paused"

// Event type to unmarshal message payloads into
type Event struct {
	Type string `json:"event"`
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
	// Make our Add Track Channel
	AddTrack = make(chan struct{}, 1)
	// Return a new Reactor
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
		return r.addTrack()
	case PAUSE_EVENT:
		return r.pausePlayer()
	case RESUME_EVENT:
		return r.resumePlayer()
	case STOP_EVENT:
		return r.stopTrack()
	}

	return nil
}

// A track was added to the player
func (r *Reactor) addTrack() error {
	log.Println("Add Event")

	if !Playing && len(AddTrack) == 0 {
		log.Println("Unblock AddTrack")
		AddTrack <- struct{}{}
	}

	return nil // always return nil since no errors can happen here
}

// Pause the Player
func (r *Reactor) pausePlayer() error {
	log.Println("Pause Event")
	// Pause the Track
	r.Player.Pause() // TODO: Use Channel
	// Set the Redis Key for Storing Player Pause State
	err := r.RedisClient.Set(PAUSE_STATE_KEY, "1", 0).Err()

	return err
}

// Resume the Player
func (r *Reactor) resumePlayer() error {
	log.Println("Resume Event")
	// Play the Track
	r.Player.Resume() // TODO: Use Channel
	// Set the Redis Key for Storing Player Pause State
	err := r.RedisClient.Set(PAUSE_STATE_KEY, "0", 0).Err()

	return err
}

// Stop the Current Track - Unloading the track causing the next track
// to be played
func (r *Reactor) stopTrack() error {
	log.Println("Stop Event")

	// Force the track to stop by placing a message on the StopTrack
	// channel which will cause the Player method to unblock
	StopTrack <- struct{}{}

	return nil // always return nil since no errors can happen here
}
