package firebase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/zabawaba99/firego"
)

type User struct {
	AvatarUrl   string `json:"avatar_url"`
	DisplayName string `json:"display_name"`
	FamilyName  string `json:"family_name"`
	GivenName   string `json:"given_name"`
	Id          string `json:"id"`
}

type QueueItem struct {
	Track Track `json:"track"`
	User  User  `json:"user"`
}

type Queue map[string]QueueItem

type Artist struct {
	Id   string `json:"id"`
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type Image struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"Height"`
}

type Album struct {
	Id     string  `json:"id"`
	Uri    string  `json:"uri"`
	Name   string  `json:"name"`
	Images []Image `json:"images"`
}

type Track struct {
	Name      string   `json:"name"`
	Uri       string   `json:"uri"`
	PlayCount int      `json:"play_count"`
	Duration  int      `json:"duration"`
	Id        string   `json:"id"`
	Album     Album    `json:"album"`
	Artist    []Artist `json:"artists"`
}

type Current struct {
	Mute    bool      `json:"mute"`
	Playing bool      `json:"playing"`
	Volume  float64   `json:"volume"`
	Item    QueueItem `json:"item"`
}

type Firebase struct {
	addr    string
	firego  *firego.Firebase
	current *firego.Firebase
	queue   *firego.Firebase
}

// Get current object form Firebase store
func (f *Firebase) GetCurrent() (*Current, error) {
	var v map[string]interface{}
	if err := f.current.Value(&v); err != nil {
		return &Current{}, err
	}

	jsonString, err := json.Marshal(v)
	if err != nil {
		return &Current{}, err
	}

	current := Current{}
	if err := json.Unmarshal(jsonString, &current); err != nil {
		return &Current{}, err
	}

	return &current, nil
}

// Get whole queue from Firebase data store
func (f *Firebase) GetQueue() (*Queue, error) {
	var v map[string]interface{}
	if err := f.queue.Value(&v); err != nil {
		return &Queue{}, err
	}

	jsonString, err := json.Marshal(v)
	if err != nil {
		return &Queue{}, err
	}

	queue := Queue{}
	if err := json.Unmarshal(jsonString, &queue); err != nil {
		return &Queue{}, err
	}

	return &queue, nil
}

// Move the first item from the queue and move it to current item. If there is
// no item in the queue current track is automatically set to empty QueueItem
func (f *Firebase) MoveToNext() error {
	queue, err := f.GetQueue()
	if err != nil {
		return err
	}
	currentTrack := f.current.Child("item")

	top := QueueItem{}
	if len(*queue) > 0 {
		keys := f.GetQueueKeys(queue)
		top = (*queue)[keys[0]]
		if err := currentTrack.Set(top); err != nil {
			return err
		}
		if err := f.queue.Child(keys[0]).Remove(); err != nil {
			return err
		}
	} else {
		if err := currentTrack.Remove(); err != nil {
			return err
		}
	}

	return nil
}

// Extract keys from queue
func (f *Firebase) GetQueueKeys(queue *Queue) []string {
	keys := make([]string, 0, len(*queue))
	for k := range *queue {
		keys = append(keys, k)
	}
	return keys
}

func (f *Firebase) WatchCurrent() {
	notifications := make(chan firego.Event)
	if err := f.current.Watch(notifications); err != nil {
		log.Fatal(err)
	}

	defer f.firego.StopWatching()
	for _ = range notifications {
		current, _ := f.GetCurrent()
		fmt.Printf("%+v\n", current)
	}
	fmt.Printf("Notifications have stopped")
}

// Factory object for firebase data store
func New(url string) *Firebase {
	f := firego.New(url, nil)
	return &Firebase{
		firego:  f,
		current: f.Child("player").Child("current"),
		queue:   f.Child("player").Child("queue"),
	}
}
