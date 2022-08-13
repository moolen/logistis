package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type Store interface {
	Observe(ev *Event) error
	List(namespace string) (map[string][]*Event, error)
}

type Event struct {
	ID        string    `json:"id"`
	Group     string    `json:"group"`
	Kind      string    `json:"kind"`
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"time"`
	Operation string    `json:"operation"`
	UserInfo  string    `json:"userInfo"`
	Object    []byte    `json:"object"`
	OldObject []byte    `json:"oldObject"`
}

func (e Event) Key() []byte {
	return []byte(fmt.Sprintf("%s/%s/%s/%s", e.Namespace, e.Group, e.Kind, e.Name))
}

func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}
