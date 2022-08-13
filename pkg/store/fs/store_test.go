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

	for _, ns := range namespaces {
		for i := 0; i < 100; i++ {
			err = st.Observe(&store.Event{
				ID:        uuid.New().String(),
				Group:     "apps",
				Kind:      "Deployment",
				Namespace: ns,
				Name:      "coredns",
				Timestamp: time.Now(),
			})
			if err != nil {
				t.Error(err)
			}
		}
	}

	events, err := st.List("kube-system")
	if err != nil {
		t.Error(err)
	}

	if len(events["kube-system/apps/Deployment/coredns"]) != 100 {
		t.Fail()
	}
}
