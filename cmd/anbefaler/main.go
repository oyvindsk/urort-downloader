package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bogem/id3v2"
	"github.com/dgraph-io/badger"

	"github.com/oyvindsk/urørt-downloader/pkg/urørt"
)

const mp3FolderF = "./musikk/%s"
const dbdir = "./badger-db"

func main() {

	// Open the DB - Once and globally (for now)
	db, err := openDB(dbdir, false)
	if err != nil {
		log.Fatalln("opening db: ", err)
	}
	defer db.Close()

	// Refresh the database. Will start from scratch if neccessary
	err = getJSONrefreshDB(db)
	if err != nil {
		log.Fatalln("getting json and refreshing database: ", err)
	}

	// Get songs recommended in 2018
	songs, err := songsRecomendedThisYear(db)
	if err != nil {
		log.Fatalln("songsRecomended: ", err)
	}

	log.Printf("Recommended in 2018: %d\n", len(songs))

	// Loop the songs and get the mp3's

	// FIXME:
	// folder structure
	// read and set "Downloaded" for each song

	for _, s := range songs {
		mp3, err := fetchMP3(s.ID)
		if err != nil {
			log.Fatal(err)
		}

		err = saveMP3WithID3(mp3, s)
		if err != nil {
			log.Fatal(err)
		}
	}

}

// getJSONrefreshDB fetch all json and store it in the database
func getJSONrefreshDB(db *badger.DB) error {

	var seenBeforeStreak int

	// handleSong() will be called for each song we find in the json, until it returns stop == true, err == true or fetchAllJSON() wants to stop
	handleSong := func(s urørt.Song) (bool, error) {

		// add the song to the db
		added, err := addSong(db, s)
		if err != nil {
			return false, err // first value should be ignored is err == true
		}

		// was it added? Or maybe it existed already?
		if added {
			seenBeforeStreak = 0
		} else {
			seenBeforeStreak++
			if seenBeforeStreak > 9 {
				log.Println("getJSONrefreshDB: saw 10 songs we have seens before, stopping JSON download")
				return true, nil
			}
		}
		return false, err
	}

	return fetchAllJSON(handleSong)
}

// getAndStoreAllJSON will fetch all json and store it in the database
// func getAndStoreAllJSON(db *badger.DB) error {
// 	return fetchAllJSON(func(s urørt.Song) error {
// 		_, err := addSong(db, s)
// 		return err
// 	})
// }

func saveMP3WithID3(mp3 io.ReadCloser, s urørt.Song) error {

	defer mp3.Close() // we got a ReadCloser, so this is our responsibility (?)

	// create the folder for the month
	p := fmt.Sprintf(mp3FolderF, s.Recommended.Format("January-2006"))
	err := os.MkdirAll(p, 0744)
	if err != nil {
		return err
	}

	// create the path
	file, err := os.Create(path.Join(p, fmt.Sprintf("%s - %s.mp3", s.BandName, s.Title)))
	if err != nil {
		return err
	}
	defer file.Close()

	// write id3v2 tag
	tag := id3v2.NewEmptyTag()
	tag.SetTitle(s.Title)
	tag.SetArtist(s.BandName)
	tag.SetYear(string(s.Released.Year())) // ?

	_, err = tag.WriteTo(file)
	if err != nil {
		return err
	}

	// write the mp3 file
	_, err = io.Copy(file, mp3)
	if err != nil {
		return err
	}

	return nil
}

func fetchMP3(songID int) (io.ReadCloser, error) {

	log.Println("fetchMP3: getting song with id:", songID)

	resp, err := http.Get(fmt.Sprintf("http://urort.p3.no/api/track/Download?trackId=%d", songID))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
