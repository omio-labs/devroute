package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	devrouter := NewDevRouter()
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go devrouter.Start()
	<-stop
	devrouter.Stop()
}
