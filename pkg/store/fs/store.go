package fs

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
)

type Store struct {
	path string
	db   *badger.DB
}

func New(path string, logger *logrus.Logger, maxVersions int) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithNumVersionsToKeep(maxVersions).WithLogger(logger))
	if err != nil {
		return nil, err
	}
	return &Store{
		path: path,
		db:   db,
	}, nil
}

func (s *Store) Observe(ev *store.Event) error {
	return s.db.Update(func(txn *badger.Txn) error {
		val, err := ev.Marshal()
		if err != nil {
			return err
		}
		entry := badger.NewEntry(ev.Key(), val)
		return txn.SetEntry(entry)
	})
}

func (s *Store) List(namespace string) (map[string][]*store.Event, error) {
	events := make(map[string][]*store.Event)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = true
		nsit := txn.NewIterator(opts)
		defer nsit.Close()
		prefix := []byte(namespace)
		for nsit.Seek(prefix); nsit.ValidForPrefix(prefix); nsit.Next() {
			item := nsit.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				fmt.Printf("key=%s, value=%s\n", k, v)
				ev := &store.Event{}
				err := ev.Unmarshal(v)
				if err != nil {
					return err
				}
				if events[string(ev.Key())] == nil {
					events[string(ev.Key())] = make([]*store.Event, 0)
				}
				events[string(ev.Key())] = append(events[string(ev.Key())], ev)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return events, err

}

func (s *Store) Close() error {
	return s.db.Close()
}
