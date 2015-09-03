// Handles connections to the Perceptor Service

package perceptor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Provides an interface to Perceptor
type Perceptor struct {
	addr   string // address to Perceptor
	secret string // soundwaves client key
}

// Generates a HMAC Signature for the given data blob
func (p *Perceptor) Sign(d []byte) []byte {
	mac := hmac.New(sha256.New, []byte(p.secret))
	mac.Write(d)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return sig
}

// Starts a websocket connection to the Perceptor Event Service
func (p *Perceptor) WSConnection() {
	// Create the dialer
	d := &websocket.Dialer{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	// Make Headers
	headers := http.Header{}
	headers.Add("Signature", fmt.Sprintf("%s:%s", "soundwave", p.Sign([]byte())))

	// Connect to the WS Service
	for {
		conn, _, err := d.Dial(fmt.Sprintf("ws://%s", p.addr), headers)
		if err != nil {
			log.Errorf("WS Dial Error: %s", err)
			time.Sleep(time.Second)
			continue
		}
	ReadLoop:
		for {
			// Read the messages, breaking on Error
			if msgType, msg, err := conn.ReadMessage(); err != nil {
				log.Errorf("WS Read Message Error: %s", e)
				conn.Close()
				break ReadLoop
			}

			// Only act on the message if it's Text
			if msgType == websocket.TextMessage {
				// TODO: Put on Chanel
				fmt.Println(msg)
			}
		}
	}

}

// Constructs a new Perceptor instance
func New() *Perceptor {
	return &Perceptor{
		addr:   a,
		secret: s,
	}
}
