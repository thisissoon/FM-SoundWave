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

// Type for holdiing Reactor configuration
type ReactorConfig struct {
	RedisChannelName string
	RedisClient      *redis.Client
	SpotifyPlayer    *spotify.Player
}

// Event reactor function. Subscribes to a Redis Pub/Sub channel
// This is a blocking method and should be called from a goroutine
func EventReactor(c *ReactorConfig) {
	// Subscribe to channel, exiting the program on fail
	pubsub := c.RedisClient.PubSub()
	err := pubsub.Subscribe(c.RedisChannelName)
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
				if m.Payload == "pause" {
					c.SpotifyPlayer.Pause()
				}
				if m.Payload == "play" {
					c.SpotifyPlayer.Play()
				}
			default:
				log.Println("Unknown message: %#v", m)
			}
		}
	}
}
