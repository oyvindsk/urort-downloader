package datastore

import "log"

type Song struct {
    UrortID                string
    FirstSeen, LastUpdated string // TODO time.Time
    Artist, Title          string
}

func (ds *Datastore) Add(song Song) error {

    _, err := ds.db.Exec(
        `INSERT INTO songs (
            urort_id,
            first_seen,
            last_updated,
            artist,
            title,
            download_started,
            download_finished
        ) VALUES ($1, datetime('now'), datetime('now'), $2, $3, false, false) 
        ON CONFLICT DO NOTHING`,

        song.UrortID,
        song.Artist,
        song.Title,
    )

    if err != nil {
        return err
    }

    return nil
}

func (ds *Datastore) DoWeHave() (bool, error) {

    return false, nil

}

func (ds *Datastore) DownloadStarted(s Song) error {

    log.Printf("DownloadStarted: %s - %s", s.Artist, s.Title)

    //     if _, err = db.Exec(`

    // insert into t values(42), (314);
    // `); err != nil {
    //         log.Fatalln(err)
    //     }

    //     rows, err := db.Query("select 3*i from t order by i;")
    //     if err != nil {
    //         log.Fatalln(err)
    //     }

    //     for rows.Next() {
    //         var i int
    //         if err = rows.Scan(&i); err != nil {
    //             log.Fatalln(err)
    //         }

    //         fmt.Println(i)
    //     }

    //     if err = rows.Err(); err != nil {
    //         log.Fatalln(err)
    //     }

    return nil
}
