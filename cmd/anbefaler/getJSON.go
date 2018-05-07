package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/oyvindsk/urørt-downloader/pkg/urørt"
)

const maxPagesToGet = 10000

func refreshDB(handleSong func(s urørt.Song) error) error {

	return nil
}

// fetchAllJSON will fetch all songs from Urørt and call the callback for each one
func fetchAllJSON(handleSong func(s urørt.Song) (bool, error)) error {

	for i := 0; i < maxPagesToGet; i++ {
		_, ur, err := getRecommendedPage(i)
		if err != nil {
			return err
		}

		log.Printf("fetchAllJSON: fetched page %d, songs: %d\n", i, len(ur.Songs))

		if len(ur.Songs) == 0 {
			log.Println("fetchAllJSON: Empty results, stopping")
			return nil
		}

		for _, v := range ur.Songs {
			stop, err := handleSong(v)
			if err != nil {
				return err
			}
			if stop {
				log.Println("fetchAllJSON: Callback returned stop. Stopping")
				return nil
			}
		}
	}
	return nil
}

// getRecommendedPage
// page starts at 0
func getRecommendedPage(page int) (*bytes.Buffer, *urørt.Page, error) {

	pageLen := 100
	skip := page * pageLen

	// http get
	resp, err := http.Get(
		fmt.Sprintf("http://urort.p3.no/breeze/urort/TrackDTOViews?$filter=Recommended%%20ne%%20null&$orderby=Recommended%%20desc%%2CId%%20desc&$top=%d&$expand=Tags%%2CFiles&$inlinecount=allpages&playlistId=0&$skip=%d", pageLen, skip),
	)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// create a tee reade to copy the results into json as well
	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	var res urørt.Page

	for {
		dec := json.NewDecoder(tee)
		if err := dec.Decode(&res); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}
	}

	return &buf, &res, nil
}
