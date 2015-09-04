//

package perceptor

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Track struct {
	Id   string `json:"uuid"`
	Uri  string `json:"uri"`
	User string `json:"user"`
}

func NewTrack(data []byte) (*Track, error) {
	t := &Track{}
	if err := json.Unmarshal(data, t); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to Unmarshal Track: %s", data))
	}

	return t, nil
}
