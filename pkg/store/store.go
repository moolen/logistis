package store

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
)

type Store interface {
	Observe(ev *Event) error
	List(namespace, kind, name string, maxHistory int) (map[string][]*Event, error)
}

type Event struct {
	ID        string                    `json:"id"`
	Group     string                    `json:"group"`
	Kind      string                    `json:"kind"`
	Namespace string                    `json:"namespace"`
	Name      string                    `json:"name"`
	Timestamp time.Time                 `json:"time"`
	Operation string                    `json:"operation"`
	UserInfo  authenticationv1.UserInfo `json:"userInfo"`
	Object    []byte                    `json:"object"`
	OldObject []byte                    `json:"oldObject"`
}

func (e Event) Key() []byte {
	return []byte(fmt.Sprintf("%s/%s/%s", strings.ToLower(e.Namespace), strings.ToLower(e.Kind), strings.ToLower(e.Name)))
}

func (e Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *Event) Unmarshal(data []byte) error {
	return json.Unmarshal(data, e)
}

func (e *Event) String() string {
	jb, _ := json.Marshal(e)
	return string(jb)
}
