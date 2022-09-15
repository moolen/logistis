package recorder

import (
	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
)

type Recorder struct {
	Logger *logrus.Logger
	store  store.Store
	match  *MatchConfig
}

func New(logger *logrus.Logger, db store.Store, match *MatchConfig) (*Recorder, error) {
	return &Recorder{
		Logger: logger,
		store:  db,
		match:  match,
	}, nil
}
