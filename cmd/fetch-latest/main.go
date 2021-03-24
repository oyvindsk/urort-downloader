package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
)

func main() {

	f, err := os.Open("scripts/test-input-1.html")
	if err != nil {
		log.Fatalln(err)
	}

	result, err := findSongs(f)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("\n\n%v\n", result)

	// Download in paralell
	eg, ctx := errgroup.WithContext(context.Background())

	for i := range result[0:3] {
		ii := i // copy it before using in the go routine
		eg.Go(func() error {
			log.Printf("Laster ned: %d", ii)
			return downloadSong("./songs", result[ii])
		})
	}

	go func() {
		select {
		// case c <- result{path, md5.Sum(data)}:
		case <-ctx.Done():
			log.Printf("CTX Done: %s", ctx.Err())
		}
	}()

	if err := eg.Wait(); err != nil {
		log.Fatalln("errorgroup returned: ", err)
	}

	// return results, nil

}

func downloadSong(toPath string, songInfo song) error {

	baseURL := "https://urort.p3.no/track/download"

	t := time.Now()

	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, songInfo.id)) // TODO use url instead?
	if err != nil {
		return err
	}

	log.Printf("\tGET took: %s", time.Since(t))

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloadSong: GET returned != 200: %s", resp.Status)
	}

	filepath := path.Join(toPath, fmt.Sprintf("%s %s - %s.wav", songInfo.id, songInfo.artist, songInfo.title))

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)

	if err != nil {
		return err
	}

	t = time.Now()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("\tio.Copy took: %s", time.Since(t))

	log.Printf("Saved %.1f MB to %q", math.Round(float64(written)/1024/1024), filepath)

	return nil
}

type song struct {
	artist, title string
	id            string
}

func findSongs(input io.Reader) ([]song, error) {

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

	var result []song

	var s song

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
					s.id = t.Attr[i].Val
					state = "lft"
					continue
				}

				log.Fatalf("div.info had no .data-trackid")
			}

		case "ina":

			if tType != html.TextToken {
				continue
			}

			s.artist = t.Data
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

			s.title = t.Data
			state = "lft"

		default:
			log.Fatalf("Unknown state: %q", state)

		}

		if s.id != "" && s.artist != "" && s.title != "" {
			result = append(result, s)
			s = song{}
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
