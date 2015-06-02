// Watches Redis List for Tracks to Play

package soundwave

import (
	"time"

	"github.com/op/go-libspotify/spotify"
	"gopkg.in/redis.v3"
)

// Type for watching the playlist queue
type Playlist struct {
	RedisKeyName   string
	RedisClient    *redis.Client
	SpotifySession *spotify.Session
}

// Constructs a New Playlist
func NewPlaylist(k string, r *redis.Client, s *spotify.Session) *Playlist {
	return &Playlist{
		RedisKeyName:   k,
		RedisClient:    r,
		SpotifySession: s,
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
			// Every second try and pop an item off the queue
		}
	}
}
