// Spotify Player

package soundwave

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/op/go-libspotify/spotify"
)

var Playing bool = false

const (
	APPLICATION_NAME  string = "SFM_"
	CACHE_LOCATION    string = "tmp"
	SETTINGS_LOCATION string = "tmp"
)

// Channel to detect when a track should be stopped
var StopTrack chan struct{}
var StopTimer chan struct{}

type Ticker struct {
	duration int
	step     int
}

func (t *Ticker) Start() {
	t.step = 1
	t.duration = 1
}

func (t *Ticker) Stop() {
	t.duration = 1
	t.step = 0
}

func (t *Ticker) Play() {
	t.step = 1
}

func (t *Ticker) Increase() {
	t.duration = t.duration + t.step
}

func (t *Ticker) Pause() {
	t.step = 0
}

func NewTicker() *Ticker {
	StopTimer = make(chan struct{}, 1)
	return &Ticker{}
}

// Soundwave player, handles holding the connection to Spotify and
// playing tracks
type Player struct {
	Audio       *audioWriter
	Session     *spotify.Session
	Player      *spotify.Player
	TrackTicker *Ticker
}

// Constructs a new player taking the Spotify user, password and key path
// as the only arguments
func NewPlayer(u *string, p *string, k *string) (*Player, error) {
	// Create a new Audio Writer, this will be used to write the audio steeam to
	audio, err := newAudioWriter()
	if err != nil {
		return nil, err // Exit on fail
	}

	// Read Key File
	key, err := ioutil.ReadFile(*k)
	if err != nil {
		return nil, err // Exit on fail
	}

	// Spotify Session Config
	config := &spotify.Config{
		ApplicationKey:   key,
		ApplicationName:  APPLICATION_NAME,
		CacheLocation:    CACHE_LOCATION,
		SettingsLocation: SETTINGS_LOCATION,
		AudioConsumer:    audio,

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	}

	// Spotify Credentials
	credentials := spotify.Credentials{
		Username: *u,
		Password: *p,
	}

	// Create Spotify Session
	session, err := spotify.NewSession(config)
	if err != nil {
		return nil, err // Exit on fail
	}

	// Login to Spotify
	if err = session.Login(credentials, true); err != nil {
		return nil, err // Exit on fail
	}

	// Set Bitrate (320kpbs)
	session.PreferredBitrate(spotify.Bitrate320k)

	// Concurrently log Session log messages
	go func() {
		for msg := range session.LogMessages() {
			log.Print(msg)
		}
	}()

	// Block until login completes, if fails exit
	select {
	case err = <-session.LoggedInUpdates():
		if err != nil {
			return nil, err // Exit on fail
		}
	}

	// Make channel to notify stop track events
	StopTrack = make(chan struct{}, 1)

	// Log connection state changes
	go WatchConnectionStateUpdates(session)

	// Create Player instance
	return &Player{
		Session:     session,
		Audio:       audio,
		Player:      session.Player(),
		TrackTicker: NewTicker(),
	}, nil

}

// Load Track from Spotify - Does not play it
func (p *Player) LoadTrack(uri *string) (*spotify.Track, error) {
	// Parse the track URI
	log.Println("Parse link:", uri)
	link, err := p.Session.ParseLink(*uri)
	if err != nil {
		return nil, err
	}

	// Get track link
	log.Println("Get Track Link")
	track, err := link.Track()
	if err != nil {
		return nil, err
	}

	// Block until the track is loaded
	log.Println("Wait for Track")
	track.Wait()

	return track, nil
}

// Play a track until the end or we get message on the StopTrack
// channel
func (p *Player) Play(uri *string) error {
	// Get the track
	log.Println("Load Track:", uri)
	track, err := p.LoadTrack(uri)
	if err != nil {
		return err
	}

	// Load the Track
	log.Println("Load Track into Player")
	if err := p.Player.Load(track); err != nil {
		return err
	}

	// Defer unloading the track until we exit this func
	defer p.Player.Unload()

	// Play the track
	log.Println(fmt.Sprintf("Playing: %s", *uri))
	p.Player.Play() // This does NOT block, we must block ourselves
	p.TrackTicker.Start()
	Playing = true

	// Runs a loop which increases track duration every second
	go p.tickerIncreaser()

	// Go routine to listen for end of track updates from the player, once we get one
	// send a message to our own StopTrack channel
	go p.EndTrack()

	<-StopTrack // Blocks
	StopTimer <- struct{}{}
	Playing = false

	log.Println(fmt.Sprintf("End: %s", *uri))

	return nil
}

// Increase ticker duration
func (p *Player) tickerIncreaser() error {
	for {
		tick := time.Tick(1 * time.Second)
		select {
		case <-StopTimer:
			return nil
		case <-tick:
			p.TrackTicker.Increase()
		}
	}
}

// Track ends
func (p *Player) EndTrack() {
	<-p.Session.EndOfTrackUpdates() // Blocks
	log.Println("End of Track Updates - Stop Track")
	p.TrackTicker.Stop()
	StopTrack <- struct{}{}
}

// Pause Track
func (p *Player) Pause() {
	p.Player.Pause()
	p.TrackTicker.Pause()
}

// Resume Track
func (p *Player) Resume() {
	p.Player.Play()
	p.TrackTicker.Play()
}

func (p *Player) IsPlaying() bool {
	return p.TrackTicker.step > 0
}

func (p *Player) CurrentElapsedTime() int {
	return p.TrackTicker.duration
}

// Watches the connection state changes with the Spotify Session and
// logs them. Note: this function blocks
//
// TODO: On changes of the connection state to anything but logged in
//       try to relogin - Unknown if this is required yet
func WatchConnectionStateUpdates(session *spotify.Session) {
	// Blocking loop, subscribes to session connection state channel
	for _ = range session.ConnectionStateUpdates() {
		var state string
		switch session.ConnectionState() {
		case 0:
			state = "Logged Out"
		case 1:
			state = "Logged In"
		case 2:
			state = "Disconnected"
		case 3:
			state = "Undefined / Unknown"
		case 4:
			state = "Offline"
		default:
			state = "Unknown State"
		}
		log.Println("Connection State:", state)
	}
}
