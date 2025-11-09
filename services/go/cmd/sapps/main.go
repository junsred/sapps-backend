package main

import (
	"flag"
	"sapps/lib/util"
	"sapps/pkg/sapps/app"
	"sapps/pkg/sapps/script"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"
)

var (
	WEBPORT = "3001"
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	script.Scripts()
	err := app.NewHTTWebPApp().Listen(":" + WEBPORT)
	if err != nil {
		log.Println("SERVER STARTED", err)
	}
	log.Println("SERVER STARTED")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Server Exited Properly")
}

func init() {
	if os.Getenv("WEBPORT") != "" {
		WEBPORT = os.Getenv("WEBPORT")
	}
	util.LoadFolder()
}
