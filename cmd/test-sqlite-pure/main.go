package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func main() {

	// dir, err := ioutil.TempDir("", "test-")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	//	defer os.RemoveAll(dir)

	fn := filepath.Join("/tmp", "test-096604565") // fn := filepath.Join(dir, "db")

	log.Printf("db filepath: %q", fn)

	db, err := sql.Open("sqlite", fn)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = db.Exec(`
	drop table if exists t;
	create table t(i);`); err != nil {
		log.Fatalln(err)
	}

	if _, err = db.Exec(`insert into t values(42), (314);`); err != nil {
		log.Fatalln(err)
	}

	rows, err := db.Query("select 3*i from t order by i;")
	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() {
		var i int
		if err = rows.Scan(&i); err != nil {
			log.Fatalln(err)
		}

		fmt.Println(i)
	}

	if err = rows.Err(); err != nil {
		log.Fatalln(err)
	}

	if err = db.Close(); err != nil {
		log.Fatalln(err)
	}

	fi, err := os.Stat(fn)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%s size: %v\n", fn, fi.Size())

}
