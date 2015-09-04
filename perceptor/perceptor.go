// Handles connections to the Perceptor Service

package perceptor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/thisissoon/FM-SoundWave/events"
)

// Provides an interface to Perceptor
type Perceptor struct {
	addr     string           // address to Perceptor
	secret   string           // soundwaves client key
	channel  chan []byte      // channel to send events too
	channels *events.Channels // Event channels
}

type playEvent struct {
	Start string `json:"start"`
	Uri   string `json:"uri"`
	User  string `json:"user"`
}

// Generates a HMAC Signature for the given data blob
func (p *Perceptor) Sign(d []byte) string {
	mac := hmac.New(sha256.New, []byte(p.secret))
	mac.Write(d)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("%s:%s", "soundwave", sig)
}

// Get the next tack from Perceptor
func (p *Perceptor) Next() (*Track, error) {
	// Build urls / client and payload
	url := fmt.Sprintf("http://%s/playlist/next", p.addr)
	client := &http.Client{}
	payload := []byte("")

	// Create Request
	req, _ := http.NewRequest("GET", url, bytes.NewBuffer(payload))
	req.Header.Add("Signature", p.Sign(payload))

	// Execute Request
	resp, err := client.Do(req)
	log.Debugf("GET %s", url)
	if err != nil {
		log.Errorf("Error getting next track: %s", err)
	}

	// Playlist is empty or errored
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Returned %v", resp.StatusCode))
	}

	// Read body and make a Track
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error getting reading next track: %v", err)
		return nil, err
	}

	t, err := NewTrack(body)
	if err != nil {
		log.Errorf("Failed making Track: %s", err)
		return nil, err
	}

	return t, nil
}

// POST's play event to perspector
func (p *Perceptor) Play(t *Track) {
	// Build urls / client
	url := fmt.Sprintf("http://%s/events/play", p.addr)
	client := &http.Client{}

	// Create payload
	payload, err := json.Marshal(&playEvent{
		Start: time.Now().UTC().Format(time.RFC3339),
		Uri:   t.Uri,
		User:  t.User})
	if err != nil {
		log.Errorf("Failed to marshal Track: %s", err)
	}

	// Create Request
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Add("Signature", p.Sign(payload))

	// Make request and log
	resp, err := client.Do(req)
	log.Infof("GET %s: %v", url, resp.StatusCode)
}

// Starts a websocket connection to the Perceptor Event Service
func (p *Perceptor) WSConnection() {
	log.Infof("Starting Websocket Connection too: %s", p.addr)

	// Create the dialer
	d := &websocket.Dialer{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	// Make Headers
	headers := http.Header{}
	headers.Add("Signature", p.Sign([]byte("")))

	// Connect to the WS Service
	for {
		conn, _, err := d.Dial(fmt.Sprintf("ws://%s", p.addr), headers)
		if err != nil {
			log.Errorf("WS Dial Error: %s", err)
			time.Sleep(time.Second)
			continue
		}
		log.Infof("Connected to: %s", p.addr)
		// Always ensure we unblock the player when we restore connections
		if len(p.channels.HasNext) == 0 {
			p.channels.HasNext <- true
		}
	ReadLoop:
		for {
			// Read the messages, breaking on Error
			if msgType, msg, err := conn.ReadMessage(); err != nil {
				log.Errorf("WS Read Message Error: %s", err)
				break ReadLoop
			} else {
				// Only act on the message if it's Text
				if msgType == websocket.TextMessage {
					p.channel <- msg
				}
			}
		}
		conn.Close()
	}
}

// Constructs a new Perceptor instance
func New(a string, s string, c chan []byte, channels *events.Channels) *Perceptor {
	return &Perceptor{
		addr:     a,
		secret:   s,
		channel:  c,
		channels: channels,
	}
}
