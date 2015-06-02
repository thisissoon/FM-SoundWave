// React to specific events from a Redis Pub/Sub channel. For example
// Pausing / Playing and Skipping the current Track.

package soundwave

import (
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

// Type for creating an event reactor
type Reactor struct {
	RedisChannelName string
	RedisClient      *redis.Client
	SpotifyPlayer    *spotify.Player
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

			default:
				log.Println("Unknown message: %#v", m)
			}
		}
	}
}

// Constructor for the Reactor type taking 3 arguments:
// - Redis Channel Name
// - Redis Client
// - Spotify Player
func NewReactor(c string, r *redis.Client, p *spotify.Player) *Reactor {
	return &Reactor{
		RedisChannelName: c,
		RedisClient:      r,
		SpotifyPlayer:    p,
	}
}
