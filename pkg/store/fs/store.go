package fs

import (
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/moolen/logistis/pkg/store"
	"github.com/sirupsen/logrus"
)

type Store struct {
	path   string
	db     *badger.DB
	logger *logrus.Logger
}

func New(path string, logger *logrus.Logger, maxVersions int) (*Store, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithNumVersionsToKeep(maxVersions).WithLogger(logger))
	if err != nil {
		return nil, err
	}
	return &Store{
		path:   path,
		db:     db,
		logger: logger,
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

func (s *Store) List(namespace, kind, name string, maxHistory int) (map[string][]*store.Event, error) {
	events := make(map[string][]*store.Event)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = true
		nsit := txn.NewIterator(opts)
		defer nsit.Close()

		// use prefix to speed up lookup
		prefix := []byte("")
		if namespace != "" {
			prefix = []byte(namespace + "/")
			if kind != "" {
				prefix = append(prefix, []byte(kind+"/")...)
				if name != "" {
					prefix = append(prefix, []byte(name+"/")...)
				}
			}
		}

		// track number of items we've fetched
		// per key
		historyMap := make(map[string]int)
		for nsit.Seek(prefix); nsit.Valid(); nsit.Next() {
			item := nsit.Item()
			k := item.Key()
			if _, ok := historyMap[string(k)]; !ok {
				historyMap[string(k)] = 0
			}
			// skip if prefix doesn't match
			if !strings.HasPrefix(string(k), namespace) {
				continue
			}
			err := item.Value(func(v []byte) error {
				if historyMap[string(k)] >= maxHistory {
					return nil
				}
				ev := &store.Event{}
				err := ev.Unmarshal(v)
				if err != nil {
					return err
				}
				if events[string(ev.Key())] == nil {
					events[string(ev.Key())] = make([]*store.Event, 0)
				}
				events[string(ev.Key())] = append(events[string(ev.Key())], ev)
				historyMap[string(k)]++
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
