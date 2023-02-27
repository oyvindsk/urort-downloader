package datastore

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type Datastore struct {
	filepath string
	db       *sql.DB
}

// OpenSQLiteDB ..
// Remmeber to close the datastore (what hapens if you don't ? dataloss? )
func OpenSQLiteDB(ffilepath string) (*Datastore, error) {

	// dir, err := ioutil.TempDir("", "test-")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer os.RemoveAll(dir)
	// fn := filepath.Join(dir, "db")

	db, err := sql.Open("sqlite", ffilepath) // filepath)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Printf("Ping failed: %s", err)
		return nil, err
	}

	// time.Sleep(time.Second * 10)

	ds := Datastore{db: db, filepath: ffilepath}

	return &ds, nil

}

func (ds *Datastore) Close() error {

	if err := ds.db.Close(); err != nil {
		return err
	}

	// fi, err := os.Stat(ds.filepath)
	// if err != nil {
	// 	return err
	// }

	// log.Printf("datastore: Closed database file: %s, size: %v", ds.filepath, fi.Size())

	return nil
}

func (ds *Datastore) CreateTablesIfNotExists() error {

	// https://sqlite.org/autoinc.html

	if _, err := ds.db.Exec(`
		CREATE TABLE IF NOT EXISTS songs (
			urort_id            INT        PRIMARY KEY NOT NULL, 	-- Not INTEGER PRIMARY KEY as that's magic in sqlite
			first_seen 			TEXT 	   NOT NULL, 				-- https://sqlite.org/quirks.html#no_separate_datetime_datatype
			last_updated		TEXT 	   NOT NULL,
			artist              TEXT       NOT NULL,
			title               TEXT       NOT NULL, 
			download_started    bool       NOT NULL, 
			download_finished   bool       NOT NULL
		)`,
	); err != nil {
		return fmt.Errorf("createTables: %s", err)
	}
	log.Printf("Table songs created")
	return nil
}
