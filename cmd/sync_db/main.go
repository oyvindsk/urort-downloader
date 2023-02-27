package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/oyvindsk/urort-downloader/internal/datastore"
	"golang.org/x/net/html"
)

func main() {

	ds, err := datastore.OpenSQLiteDB("songs.sqlite")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		log.Printf("defer closed db")
		if err := ds.Close(); err != nil {
			log.Printf("\t with error: %s", err)
		}
	}()

	// Ensure the table(s) exist
	err = ds.CreateTablesIfNotExists()
	if err != nil {
		log.Fatalln(err)
	}

	err = get(ds)
	if err != nil {
		log.Fatalln(err)
	}

}

func get(ds *datastore.Datastore) error {

	f, err := os.Open("scripts/test-input-1.html")
	if err != nil {
		return err
	}

	result, err := findSongsInHTML(f)
	if err != nil {
		return err
	}

	log.Printf("\n\n%v\n", result)

	for _, s := range result {
		// err = ds.DownloadStarted(s)
		err = ds.Add(s)
		if err != nil {
			return err
		}
	}

	return nil

}

// FIXME: log.Fatals
func findSongsInHTML(input io.Reader) ([]datastore.Song, error) {

	/*

		States at the beginning of the big for loop:

		Current State								Input that switches state 			Next state
		----------------------------------------------------------------------------------------------------------
		(lft)	Looking for div            			StartTag div, with class in .. 		ina or int
		(ina)	In Artist 							TextToken							lft
		(int) 	In Title 							StartTag a 							inta
		(inta)	In Title -> A 						TextToken							lft


		We are looking for one of:
			- <div class="info" data-trackid="191631" data-trackurl="https://nrk-urort-prod.s3.amazonaws.com/tracks/5a8434f6-a157-4cb4-b63b-fa6075895b9c">
			- <div class="artist"><a href="/artist/isabelle-eberdean">Isabelle Eberdean</a></div>
			- <div class="title"> <a href="/track/isabelle-eberdean/focus-on-you">Focus on You</a> </div>

	*/

	z := html.NewTokenizer(input)

	state := "lft"

	var result []datastore.Song

	var s datastore.Song

	var err error

	for {

		// Advance to next token
		tType := z.Next()
		if tType == html.ErrorToken {
			// This includes EOF, break out and deal with it later
			err = z.Err()
			break // MACHINE
		}

		t := z.Token() // The token we are currenlty looking at

		//
		//

		// log.Printf("%q\t%q\t%q\n", tType, t.Type.String(), t.Data)

		switch state {

		case "lft":

			if tType != html.StartTagToken || t.Data != "div" {
				continue
			}

			if ok, _ := findAttrVal(t.Attr, "class", "artist"); ok {
				state = "ina"
				continue
			}

			if ok, _ := findAttrVal(t.Attr, "class", "title"); ok {
				state = "int"
				continue
			}

			if ok, _ := findAttrVal(t.Attr, "class", "info"); ok {

				if ok, i := findAttr(t.Attr, "data-trackid"); ok {
					s.UrortID = t.Attr[i].Val
					state = "lft"
					continue
				}

				log.Fatalf("div.info had no .data-trackid")
			}

		case "ina":

			if tType != html.TextToken {
				continue
			}

			s.Artist = t.Data
			state = "lft"

		case "int":

			if tType != html.StartTagToken || t.Data != "a" {
				continue
			}
			state = "inta"

		case "inta":

			if tType != html.TextToken {
				continue
			}

			s.Title = t.Data
			state = "lft"

		default:
			log.Fatalf("Unknown state: %q", state)

		}

		if s.UrortID != "" && s.Artist != "" && s.Title != "" {
			result = append(result, s)
			s = datastore.Song{}
		}

	}

	// Any parse / state machine error from?
	if err != nil {
		if err != io.EOF {
			log.Fatalln(fmt.Errorf("InsertContactForm: error when running state machine: %s", err))
		}
	}

	return result, nil
}

// package internal helper funcs

func findAttr(attrs []html.Attribute, key string) (bool, int) {
	for i := range attrs {
		if attrs[i].Key == key {
			return true, i // assume only 1 match
		}
	}
	return false, 0
}

func findAttrVal(attrs []html.Attribute, key, val string) (bool, int) {
	for i := range attrs {
		if attrs[i].Key == key {
			if attrs[i].Val == val {
				return true, i // assume only 1 match
			}
		}
	}
	return false, 0
}
