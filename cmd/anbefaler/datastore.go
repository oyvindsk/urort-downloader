package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/oyvindsk/urørt-downloader/pkg/urørt"
)

//
// General, DB level, functions

func openDB(dir string, readOnly bool) (*badger.DB, error) {
	// Open the Badger database located in dir. It will be created if it doesn't exist.
	opts := badger.DefaultOptions
	if readOnly {
		opts.ReadOnly = true
	}
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
	// caller must defer db.Close()
}

//
// DB Get and Set functions

// unused, untested:
// func songExists(db *badger.DB, songID int) (bool, error) {
// 	k := []byte(fmt.Sprintf("song-%d", songID))
// 	var exists bool
// 	err := db.View(func(txn *badger.Txn) error {
// 		// does it exists?
// 		_, err := txn.Get(k)
// 		if err == nil {
// 			exists = true
// 			return nil
// 		} else if err != badger.ErrKeyNotFound {
// 			return err
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return false, err
// 	}
// 	return exists, nil
// }

func addSong(db *badger.DB, s urørt.Song) (bool, error) {

	k := []byte(fmt.Sprintf("song-%d", s.ID))

	var added bool

	// Write to badger db
	err := db.Update(func(txn *badger.Txn) error {

		// does it exists?
		_, err := txn.Get(k)
		if err == nil {
			log.Printf("addSong: %s exists, skipping!", k)
			return nil
		} else if err != badger.ErrKeyNotFound {
			return err
		}

		j, err := s.ToJSON()
		if err != nil {
			return err
		}

		added = true // new song, we added it

		return txn.Set([]byte(k), j)
	})

	if err != nil {
		return false, err
	}

	return added, nil
}

func getSongDownloaded(db *badger.DB, songID int) (time.Time, error) {

	k := []byte(fmt.Sprintf("song-%d", songID))

	var downloaded time.Time

	// Write to badger db
	err := db.View(func(txn *badger.Txn) error {

		// does it exists? It should!
		i, err := txn.Get(k)
		if err != nil {
			return err
		}

		// get json bytes
		j, err := i.Value()
		if err != nil {
			return err
		}

		// unmarshall json
		s, err := urørt.SongFromJSON(j)
		if err != nil {
			return err
		}

		downloaded = s.Downloaded
		return nil
	})

	if err != nil {
		return time.Time{}, err
	}

	return downloaded, nil
}

func setSongDownloaded(db *badger.DB, songID int) error {

	k := []byte(fmt.Sprintf("song-%d", songID))

	// Write to badger db
	return db.Update(func(txn *badger.Txn) error {

		// does it exists? It should!
		i, err := txn.Get(k)
		if err != nil {
			return err
		}

		// get json bytes
		j, err := i.Value()
		if err != nil {
			return err
		}

		// unmarshall json
		s, err := urørt.SongFromJSON(j)
		if err != nil {
			return err
		}

		// update
		s.Downloaded = time.Now()

		// marshall to json again
		j, err = s.ToJSON()
		if err != nil {
			return err
		}

		// store and return
		return txn.Set([]byte(k), j)
	})
}

//
// Filters etc
func songsRecomended2018(db *badger.DB) ([]urørt.Song, error) {

	var songs []urørt.Song

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			v, err := item.Value()
			if err != nil {
				return err
			}

			song, err := urørt.SongFromJSON(v)
			if err != nil {
				return err
			}

			t, err := time.Parse(time.RFC3339, "2018-01-01T00:00:00Z") // FIXME timzone? find "this year" :D
			if err != nil {
				return err
			}

			if song.Recommended.After(t) {
				songs = append(songs, *song)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return songs, nil
}
