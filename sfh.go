package main

import "net/http"
import (
	"github.com/gorilla/mux"
	"fmt"
	"flag"
	"log"
	"os"
	"bufio"
)

var file *string
var content []string

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", File).Methods(http.MethodGet, http.MethodOptions)

	var port = flag.Int("port", 1234, "port")
	file = flag.String("file", "www.txt", "file")
	var cache = flag.Bool("cache", true, "cache")

	flag.Parse()

	if *cache {
		var err error
		content, err = readFile()
		if (err!= nil) {
			panic(err)
		}
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}

func readFile() ([]string, error) {
	f, err := os.OpenFile(*file, os.O_RDONLY, 0644)
	if (err != nil) {
		return nil, err
	}
	defer f.Close()
	res := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	return res, nil
}

func File(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" {
		w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	} else {
		w.Header().Add("Access-Control-Allow-Origin", "*")
	}

	if r.Method == http.MethodOptions {
		w.Header().Add("Access-Control-Allow-Method", "*")
		w.WriteHeader(http.StatusOK)
		return
	}

	cnt := content
	if content == nil {
		var err error
		cnt, err = readFile()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Add("Content-type", "text/html")
	for _, s := range cnt {
		w.Write([]byte(s))
	}
}
