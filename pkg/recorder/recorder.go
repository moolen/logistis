package recorder

import (
	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
)

type Recorder struct {
	Logger *logrus.Logger
	store  store.Store
}

func New(logger *logrus.Logger, db store.Store) (*Recorder, error) {
	return &Recorder{
		Logger: logger,
		store:  db,
	}, nil
}
