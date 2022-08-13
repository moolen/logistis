package fs

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
)

func TestStore(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmp)
	st, err := New(tmp, logrus.New(), 100)
	if err != nil {
		t.Error(err)
	}

	defer st.Close()

	namespaces := []string{"kube-system", "default", "example"}
	kinds := []string{"Deployment", "Statefulset", "DaemonSet"}
	names := []string{"foo", "bar", "coredns"}

	for _, ns := range namespaces {
		for _, kind := range kinds {
			for _, name := range names {
				for i := 0; i < 100; i++ {
					err = st.Observe(&store.Event{
						ID:        uuid.New().String(),
						Group:     "noop",
						Kind:      kind,
						Namespace: ns,
						Name:      name,
						Timestamp: time.Now(),
					})
					if err != nil {
						t.Error(err)
					}
				}
			}
		}
	}

	events, err := st.List("kube-system", "Deployment", "", 3)
	if err != nil {
		t.Error(err)
	}

	t.Logf("events: %#v", events)
	if len(events["kube-system/Deployment/coredns"]) != 3 {
		t.Fail()
	}
	if len(events["default/Deployment/coredns"]) != 3 {
		t.Fail()
	}
	if len(events["example/Deployment/coredns"]) != 3 {
		t.Fail()
	}
	t.Fail()
}
