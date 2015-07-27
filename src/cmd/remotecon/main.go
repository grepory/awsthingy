package main

import (
	"remotecon/api"
	"os"
	"os/signal"
)

func main() {
	go api.ListenAndServe("0.0.0.0:8080", os.Stdout)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
}
