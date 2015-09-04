// Channls that are used to send events to the player

package events

type Channels struct {
	Add     chan []byte
	Play    chan []byte
	End     chan []byte
	Pause   chan []byte
	Resume  chan []byte
	Stop    chan bool
	HasNext chan bool
}

func NewChannels() *Channels {
	return &Channels{
		Add:     make(chan []byte),
		Play:    make(chan []byte),
		End:     make(chan []byte),
		Pause:   make(chan []byte),
		Resume:  make(chan []byte),
		Stop:    make(chan bool),
		HasNext: make(chan bool, 1),
	}
}
