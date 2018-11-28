package main

import (
	"flag"
	"github.com/aaronland/go-storage"
	"log"
)

func main(){

	str_dsn := flag.String("dsn", "root=.", "...")
	fname := flag.String("filename", "echo.txt", "...")	

	flag.Parse()

	store, err := storage.NewFSStore(*str_dsn)

	if err != nil {
		log.Fatal(err)
	}

	key := *fname
	
	fh, err := store.Open(key)

	if err != nil {
		log.Fatal(err)
	}

	for _, str := range flag.Args() {
		fh.Write([]byte(str + "\n"))
	}

	fh.Close()

	log.Println(store.URI(key))
}
