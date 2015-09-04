// Handles setting up and running the Spotify Player

package player

import (
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/op/go-libspotify/spotify"
	"github.com/thisissoon/FM-SoundWave/events"
	"github.com/thisissoon/FM-SoundWave/perceptor"
)

// Vars for holding pause state
var (
	PAUSE_DURATION int64     = 0 // Total time we have been paused
	PAUSE_START    time.Time     // Time time current pause was started
)

// Our Actual Spotify Player
type Player struct {
	audio    *audioWriter
	session  *spotify.Session
	player   *spotify.Player
	pcptr    *perceptor.Perceptor
	channels *events.Channels
}

// Runs the player - plays the sweet sweet music
func (p *Player) Run() {
	for {
		<-p.channels.HasNext // Block until we have a next track
		track, err := p.pcptr.Next()
		if err != nil {
			log.Infof("Failed to Get Track: %s", err)
			continue
		}
		p.play(track)      // Blocks
		p.pcptr.End(track) // Publish end event
	}
}

// Handles recieving add events
func (p *Player) addEventHandler() {
	for {
		<-p.channels.Add
		log.Debug("Handle Add Event")
		if len(p.channels.HasNext) == 0 {
			p.channels.HasNext <- true
		}
	}
}

// Handles pause events
func (p *Player) pauseEventHandler() {
	for {
		pause := <-p.channels.Pause
		if pause {
			log.Info("Pause Player")
			PAUSE_START = time.Now().UTC()
			go p.pcptr.Pause(PAUSE_START)
			p.player.Pause()
		} else {
			log.Info("Resume Player")
			now := time.Now().UTC()
			delta := now.Sub(PAUSE_START)
			paused_for := delta.Nanoseconds() / int64(time.Millisecond)
			PAUSE_DURATION += paused_for
			go p.pcptr.Resume(PAUSE_DURATION)
			p.player.Play()
		}
	}
}

// Load Track from Spotify - Does not play it
func (p *Player) loadTrack(uri string) (*spotify.Track, error) {
	log.Infof("Load Track: %s", uri)

	// ParsePrintln the track URI
	log.Debug("Parse link:", uri)
	link, err := p.session.ParseLink(uri)
	if err != nil {
		return nil, err
	}

	// Get track link
	log.Debug("Get Track Link")
	track, err := link.Track()
	if err != nil {
		return nil, err
	}

	// Block until the track is loaded
	log.Debug("Wait for Track")
	track.Wait()

	return track, nil
}

// Play a track until the end or we get message on the StopTrack channel
func (p *Player) play(t *perceptor.Track) error {
	// Reset Pause State
	PAUSE_DURATION = 0
	PAUSE_START = time.Time{}

	// Get the track
	track, err := p.loadTrack(t.Uri)
	if err != nil {
		return err
	}

	// Load the Track
	log.Info("Load Track into Player")
	if err := p.player.Load(track); err != nil {
		return err
	}

	// Defer unloading the track until we exit this func
	defer p.player.Unload()

	// Send play event to perspector - go routine so we don't block
	go func() {
		p.pcptr.Play(t, time.Now().UTC())
		return
	}()

	// Play the track
	log.Println(fmt.Sprintf("Playing: %s", t.Uri))
	p.player.Play() // This does NOT block, we must block ourselves

	// Go routine to listen for end of track updates from the player, once we get one
	// send a message to our own StopTrack channel
	go func() {
		<-p.session.EndOfTrackUpdates() // Blocks
		log.Debug("End of Track Updates")
		p.channels.Stop <- true
		return
	}()

	<-p.channels.Stop // Blocks
	log.Infof(fmt.Sprintf("Track stopped: %s", t.Uri))

	return nil
}

// Constructs a new Spotify Player instance
func New(
	user string,
	pass string,
	keyPath string,
	pcptr *perceptor.Perceptor,
	channels *events.Channels) (*Player, error) {

	var err error

	// Create a new Audio Writer, this will be used to write the audio steeam to
	log.Debug("Spotify: Create Audio Writter")
	audio, err := newAudioWriter()
	if err != nil {
		return nil, err // Exit on fail
	}

	// Read Key File
	log.Debug("Spotify: Read Key")
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err // Exit on fail
	}

	// Spotify Credentials
	log.Debug("Spotify: Create Credentials")
	creds := spotify.Credentials{
		Username: user,
		Password: pass,
	}

	// Make a ASession
	log.Debug("Spotify: Create Session")
	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   key,
		ApplicationName:  APPLICATION_NAME,
		CacheLocation:    CACHE_LOCATION,
		SettingsLocation: SETTINGS_LOCATION,
		AudioConsumer:    audio,

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	})
	if err != nil {
		return nil, err // Exit on fail
	}

	// Log Session Events
	go func() {
		for msg := range session.LogMessages() {
			log.Debugf("Session: %s", msg)
		}
	}()

	// Set Bitrate (320kpbs)
	log.Debugf("Spotify: Set Preferred Bitrate")
	session.PreferredBitrate(BITRATE)

	// Login
	if err = session.Login(creds, true); err != nil {
		return nil, err // Exit on fail
	}

	// Make the player
	player := &Player{
		audio:    audio,
		session:  session,
		pcptr:    pcptr,
		channels: channels,
		player:   session.Player(),
	}

	// Start our event handlers
	go player.addEventHandler()
	go player.pauseEventHandler()

	return player, nil
}
