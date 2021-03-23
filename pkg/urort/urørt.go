package urort

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

//  https://mholt.github.io/json-to-go/

// Page is the resposne for 1 "page" from urørt. Contains 0 to many songs in Songs ("Results")
type Page struct {
	ID          string `json:"$id"`
	Type        string `json:"$type"`
	Songs       []Song `json:"Results"`
	InlineCount int    `json:"InlineCount"`
}

// ToJSON returns a Page as JSON, for storag etc
func (p Page) ToJSON() ([]byte, error) {
	return json.Marshal(&p)
}

// Song is a song from urørt + "Downloaded" that we set
// TODO: separete urørt data from our data?
type Song struct {
	IDs                string        `json:"$id"`
	Type               string        `json:"$type"`
	ID                 int           `json:"Id"`
	Title              string        `json:"Title"`
	Composer           string        `json:"Composer"`
	Songwriter         string        `json:"Songwriter"`
	Released           time.Time     `json:"Released"`
	TrackState         int           `json:"TrackState"`
	Recommended        time.Time     `json:"Recommended"`
	LikeCount          int           `json:"LikeCount"`
	PlayCount          int           `json:"PlayCount"`
	BandID             int           `json:"BandId"`
	BandName           string        `json:"BandName"`
	InternalBandURL    string        `json:"InternalBandUrl"`
	Image              string        `json:"Image"`
	PlayedOnRadioCount int           `json:"PlayedOnRadioCount"`
	IsPlayable         bool          `json:"IsPlayable"`
	Tags               []interface{} `json:"Tags"`
	Files              []struct {
		IDs      string `json:"$id"`
		Type     string `json:"$type"`
		ID       int    `json:"Id"`
		TrackID  int    `json:"TrackId"`
		FileRef  string `json:"FileRef"`
		FileType string `json:"FileType"`
	} `json:"Files"`
	CommentCount int `json:"CommentCount,omitempty"`

	Downloaded time.Time // Used to remember if we have downloaded the mp3 or not
}

// ToJSON returns a Song as JSON, for storag etc
func (ur Song) ToJSON() ([]byte, error) {
	return json.Marshal(&ur)
}

func (ur Song) String() string {

	buf := bytes.NewBufferString(fmt.Sprintf(`
	Downloaded %s

	ID                 %d
	IDs                %s
	Type               %s
	Title              %s
	Composer           %s
	Songwriter         %s
	Released           %s
	TrackState         %d
	Recommended        %s
	LikeCount          %d
	PlayCount          %d
	BandID             %d
	BandName           %s
	InternalBandURL    %s
	Image              %s
	PlayedOnRadioCount %d
	IsPlayable         %v
	CommentCount       %d


	Tags               %+v
	Files:
	`,
		ur.Downloaded,

		ur.ID,
		ur.IDs,
		ur.Type,
		ur.Title,
		ur.Composer,
		ur.Songwriter,
		ur.Released,
		ur.TrackState,
		ur.Recommended,
		ur.LikeCount,
		ur.PlayCount,
		ur.BandID,
		ur.BandName,
		ur.InternalBandURL,
		ur.Image,
		ur.PlayedOnRadioCount,
		ur.IsPlayable,
		ur.CommentCount,
		ur.Tags,
	))

	for _, v := range ur.Files {
		_, err := fmt.Fprintf(buf, `
			ID 			%d
			IDs			%s
			Type     	%s
			TrackID  	%d
			FileRef 	%s  
			FileType  	%s
			
			`,
			v.ID,
			v.IDs,
			v.Type,
			v.TrackID,
			v.FileRef,
			v.FileType,
		)
		if err != nil {
			log.Printf("urørt: String: %s", err)
		}
	}

	return buf.String()

}

// SongFromJSON returns a Song from a song JSON blob. Typically used when fetching from the db
// TODO call ..UnmarshallJSON() instead? Like time..  ??
func SongFromJSON(j []byte) (*Song, error) {

	var s Song
	err := json.Unmarshal(j, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
