// Watches Redis List for Tracks to Play

package soundwave

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/op/go-libspotify/spotify"

	"gopkg.in/redis.v3"
)

// Redis Keys
const (
	CURRENT_KEY                string = "fm:player:current"
	CURRENT_TRACK_ELAPSED_TIME string = "fm:player:elapsed_time"
	START_TIME_KEY             string = "fm:player:start_time"
	PAUSE_TIME_KEY             string = "fm:player:pause_time"
	PAUSED_DURATION_KEY        string = "fm:player:pause_duration"
)

// Event Names
const (
	PLAY_EVENT string = "play"
	END_EVENT  string = "end"
)

// Type for creating a play event message to publish to redis, JSON structure is
// as follows:
//
// {"event": "play", "uri": "spotify:track:1234", "user": "1234"}
type PublishEvent struct {
	Event string `json:"event"`
	Uri   string `json:"uri"`
	User  string `json:"user"`
}

// Type for unmarshaling a playlist item JSON string
type PlaylistItem struct {
	Uri  string `json:"uri"`
	User string `json:"user"`
}

// Type for watching the playlist queue
type Playlist struct {
	RedisKeyName     string
	RedisChannelName string
	RedisClient      *redis.Client
	Player           *Player
}

// Constructs a New Playlist
func NewPlaylist(k string, c string, r *redis.Client, p *Player) *Playlist {
	return &Playlist{
		RedisKeyName:     k,
		RedisChannelName: c,
		RedisClient:      r,
		Player:           p,
	}
}

// Watch the Queue checking the queue every second attempting to
// pop and item of it, once a track has been popped of the queue
// the track can be played, this should block the tick until the
// player unblocks itself either when the track finishes playing
// or the track is stopped
func (p *Playlist) Watch() {
	for {
		if p.Player.Session.ConnectionState() == spotify.ConnectionStateLoggedIn {
			// Get the next track of the queue
			track, err := p.next() // Blocks until we get an item on the queue

			// We got an err from Redis, lets just log it
			if err != nil {
				log.Println(err)
			} else {
				if err := p.prefetch(); err != nil {
					log.Println(err)
				}
				p.play(track) // Blocks
			}
		}
	}
}

// Pre-Fetch the next track if we have one
func (p *Playlist) prefetch() error {
	var err error

	// Get next item in list, returns a list
	result, err := p.RedisClient.LRange(p.RedisKeyName, 1, 1).Result()
	if err != nil {
		return err
	}

	// If we have a next track then lets prefetch it whilst the current one plays
	if len(result) == 1 {

		// Decode the JSON
		i := &PlaylistItem{}
		err = json.Unmarshal([]byte(result[0]), i)
		if err != nil {
			return err
		}

		// Tell player to prefetch
		err = p.Player.Prefetch(&i.Uri)
		if err != nil {
			return err
		}
	}

	return nil
}

// Pop track of the top of the queue returning the value of the key or nil
func (p *Playlist) next() (string, error) {
	// Value will be a []string containing [key, value]
	result, err := p.RedisClient.BLPop(0, p.RedisKeyName).Result() // Blocks
	if err == redis.Nil {
		return "", nil // No key so no queue
	} else if err != nil {
		return "", err
	} else {
		return result[1], nil
	}
}

// Publish current tract duration into redis
func (p *Playlist) CurrentTrackDurationPublisher() {
	for {
		tick := time.Tick(1 * time.Second)
		select {
		case <-tick:
			duration := p.Player.CurrentElapsedTime()
			if p.Player.IsPlaying() {
				p.RedisClient.Set(CURRENT_TRACK_ELAPSED_TIME, strconv.Itoa(duration), 0).Err()
			} else {
				p.RedisClient.Del(CURRENT_TRACK_ELAPSED_TIME)
			}
		}
	}
}

// Play the track popped of the list, this unmarshales the JSON
// value to get the track uri and pass it to the player to play
func (p *Playlist) play(value string) {
	item := &PlaylistItem{}
	err := json.Unmarshal([]byte(value[:]), item)
	if err != nil {
		log.Println(err)
	} else {
		// Publish Play Event
		if err := p.publishPlayEvent(item); err != nil {
			log.Println(err)
		} else {
			// Play the Track
			if err := p.Player.Play(&item.Uri); err != nil { // Blocks
				log.Println(err)
			} else {
				// Publish End Event
				if err := p.publishEndEvent(item); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

// Publishes a Play event to a Redis channel and also sets the value
// for the fm:player:current key
func (p *Playlist) publishPlayEvent(item *PlaylistItem) error {
	log.Println("Publish Play Event")

	var err error

	// Generate Current JSON payload
	current, err := json.Marshal(&PlaylistItem{
		Uri:  item.Uri,
		User: item.User,
	})
	if err != nil {
		return err
	}

	// Set Current Track
	err = p.RedisClient.Set(CURRENT_KEY, string(current[:]), 0).Err()
	if err != nil {
		return err
	}

	// Set Start Time
	now := time.Now()
	err = p.RedisClient.Set(START_TIME_KEY, now.Format(time.RFC3339), 0).Err()
	if err != nil {
		return err
	}

	// Generate message JSON Payload
	message, err := json.Marshal(&PublishEvent{
		Event: PLAY_EVENT,
		Uri:   item.Uri,
		User:  item.User,
	})
	if err != nil {
		return err
	}

	// Publish Message
	err = p.RedisClient.Publish(p.RedisChannelName, string(message[:])).Err()
	if err != nil {
		return err
	}

	return nil
}

// Publish end event to a Redis Channel and also delete the fm:player:current key
func (p *Playlist) publishEndEvent(item *PlaylistItem) error {
	log.Println("Publish End Event")

	var err error

	// Delete Current Track Key
	err = p.RedisClient.Del(CURRENT_KEY).Err()
	if err != nil {
		return err
	}

	// Delete Start Time Key
	err = p.RedisClient.Del(START_TIME_KEY).Err()
	if err != nil {
		return err
	}

	// Generate message JSON Payload
	message, err := json.Marshal(&PublishEvent{
		Event: END_EVENT,
		Uri:   item.Uri,
		User:  item.User,
	})
	if err != nil {
		return err
	}

	// Publish Message
	err = p.RedisClient.Publish(p.RedisChannelName, string(message[:])).Err()
	if err != nil {
		return err
	}

	return nil
}
