package main

import "net/http"
import (
	"github.com/gorilla/mux"
	"github.com/sevlyar/go-daemon"
	"fmt"
	"flag"
	"log"
	"os"
	"bufio"
	"syscall"
)

var file *string
var content []string

func main() {

	var logf = flag.String("log", "sfh.log", "log")
	var pid = flag.String("pid", "sfh.pid", "pid")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var port = flag.Int("port", 1234, "port")
	file = flag.String("file", "www.txt", "file")
	var cache = flag.Bool("cache", true, "cache")
	flag.Parse()

	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)

	cntxt := &daemon.Context{
		PidFileName: *pid,
		PidFilePerm: 0644,
		LogFileName: *logf,
		LogFilePerm: 0640,
		WorkDir:     "/tmp",
		Umask:       027,
		Args:        os.Args,
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon:", err)
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalln(err)
	}
	if d != nil {
		return
	}

	flag.Parse()
	defer cntxt.Release()

	log.Println("daemon started")

	go func() {
		router := mux.NewRouter()
		router.HandleFunc("/", File).Methods(http.MethodGet, http.MethodOptions)

		if *cache {
			var err error
			content, err = readFile()
			if (err!= nil) {
				log.Panic(err)
			}
		}
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
	}()

	err = daemon.ServeSignals()
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("daemon terminated")

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

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}