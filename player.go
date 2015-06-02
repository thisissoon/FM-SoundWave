// Spotify Player

package soundwave

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"

	"github.com/op/go-libspotify/spotify"
)

var StopTrack chan struct{}

func NewSession(user *string, pass *string, key *string) (*spotify.Session, *audioWriter) {
	debug := true

	appKey, err := ioutil.ReadFile(*key)
	if err != nil {
		log.Fatal(err)
	}

	var silenceStderr = DiscardFd(syscall.Stderr)
	if debug == true {
		silenceStderr.Restore()
	}

	audio, err := newAudioWriter()
	if err != nil {
		log.Fatal(err)
	}
	silenceStderr.Restore()

	session, err := spotify.NewSession(&spotify.Config{
		ApplicationKey:   appKey,
		ApplicationName:  "SOON_ FM",
		CacheLocation:    "tmp",
		SettingsLocation: "tmp",
		AudioConsumer:    audio,

		// Disable playlists to make playback faster
		DisablePlaylistMetadataCache: true,
		InitiallyUnloadPlaylists:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	credentials := spotify.Credentials{
		Username: *user,
		Password: *pass,
	}
	if err = session.Login(credentials, true); err != nil {
		log.Fatal(err)
	}

	// Set Bitrate
	session.PreferredBitrate(spotify.Bitrate320k)

	// Log messages
	if debug {
		go func() {
			for msg := range session.LogMessages() {
				log.Print(msg)
			}
		}()
	}

	// Wait for login and expect it to go fine
	select {
	case err = <-session.LoggedInUpdates():
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Session Created")

	StopTrack = make(chan struct{}, 1)

	return session, audio
}

func LoadTrack(session *spotify.Session, id *string) *spotify.Track {
	uri := fmt.Sprintf("spotify:track:%s", *id)
	log.Println(uri)

	// Parse the track
	link, err := session.ParseLink(uri)
	if err != nil {
		log.Fatal(err)
	}
	track, err := link.Track()
	if err != nil {
		log.Fatal(err)
	}

	// Load the track and play it
	track.Wait()

	return track
}

func Play(session *spotify.Session, player *spotify.Player, track *spotify.Track) {
	if err := player.Load(track); err != nil {
		fmt.Println("%#v", err)
		log.Fatal(err)
	}

	defer player.Unload()

	log.Println("Playing...")
	player.Play()

	// Go routine to listen for end of track updates from the player, once we get one
	// send a message to our own StopTrack channel
	go func() {
		<-session.EndOfTrackUpdates()
		log.Println("End of Track Updates - Stop Track")
		StopTrack <- struct{}{}
	}()

	<-StopTrack // Blocks

	// Unload the Track
	player.Unload()
	log.Println("End of Track")
}
