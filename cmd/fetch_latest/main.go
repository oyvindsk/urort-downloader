package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/oyvindsk/urort-downloader/internal/datastore"
)

func main() {
	err := foo()
	if err != nil {
		log.Fatalf("Shutting down with error: %s", err)
	}
	log.Println("Shutting down")
}

func foo() error {

	// // Download in paralell
	// eg, ctx := errgroup.WithContext(context.Background())

	// for i := range result[0:3] {
	// 	ii := i // copy it before using in the go routine
	// 	eg.Go(func() error {
	// 		log.Printf("Laster ned: %d", ii)
	// 		return downloadSong("./songs", result[ii])
	// 	})
	// }

	// go func() {
	// 	select {
	// 	// case c <- result{path, md5.Sum(data)}:
	// 	case <-ctx.Done():
	// 		log.Printf("CTX Done: %s", ctx.Err())
	// 	}
	// }()

	// if err := eg.Wait(); err != nil {
	// 	log.Fatalln("errorgroup returned: ", err)
	// }

	// return results, nil

	// if err := ds.Close(); err != nil {
	// 	return err
	// }

	return nil
}

func downloadSong(toPath string, songInfo datastore.Song) error {

	baseURL := "https://urort.p3.no/track/download"

	t := time.Now()

	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, songInfo.UrortID)) // TODO use url instead?
	if err != nil {
		return err
	}

	log.Printf("\tGET took: %s", time.Since(t))

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloadSong: GET returned != 200: %s", resp.Status)
	}

	filepath := path.Join(toPath, fmt.Sprintf("%s %s - %s.wav", songInfo.UrortID, songInfo.Artist, songInfo.Title))

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
