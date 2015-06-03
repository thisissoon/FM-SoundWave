// Watches Redis List for Tracks to Play

package soundwave

import (
	"encoding/json"
	"log"
	"time"

	"github.com/op/go-libspotify/spotify"

	"gopkg.in/redis.v3"
)

const CURRENT_KEY string = "fm:player:current"

// Type for unmarshaling a playlist item JSON string
type PlaylistItem struct {
	Uri  string `json:"uri"`
	User string `json:"user"`
}

// Type for watching the playlist queue
type Playlist struct {
	RedisKeyName string
	RedisClient  *redis.Client
	Player       *Player
}

// Constructs a New Playlist
func NewPlaylist(k string, r *redis.Client, p *Player) *Playlist {
	return &Playlist{
		RedisKeyName: k,
		RedisClient:  r,
		Player:       p,
	}
}

// Watch the Queue checking the queue every second attempting to
// pop and item of it, once a track has been popped of the queue
// the track can be played, this should block the tick until the
// player unblocks itself either when the track finishes playing
// or the track is stopped
func (p *Playlist) Watch() {
	for {
		tick := time.Tick(1 * time.Second)
		select {
		case <-tick:
			// If we are not logged in then we won't try and play a tack, we will try
			// again on the next tick
			if p.Player.Session.ConnectionState() == spotify.ConnectionStateLoggedIn {
				value, err := p.RedisClient.LPop(p.RedisKeyName).Result()
				if err == redis.Nil {
					// Key does not exist so no items on the queue, no need to log this, would be
					// very vebose
				} else if err != nil {
					log.Println(err)
				} else {
					// Play the item we just popped off the list
					err := p.RedisClient.Set(CURRENT_KEY, value, 0).Err()
					if err != nil {
						log.Println(err)
					} else {
						p.Play(value) // Blocks
					}
				}
			}
		}
	}
}

//
func (p *Playlist) Play(value string) {
	item := &PlaylistItem{}
	err := json.Unmarshal([]byte(value[:]), item)
	if err != nil {
		log.Println(err)
	} else {
		if err := p.Player.Play(&item.Uri); err != nil { // Blocks
			log.Println(err)
		}
	}
}
