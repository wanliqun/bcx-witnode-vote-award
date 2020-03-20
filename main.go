package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"

	"github.com/wanliqun/bcx-witnode-vote-award/cmd"
	"github.com/wanliqun/bcx-witnode-vote-award/action"
)

func main() {
	hookInterruptHandler()

	cmd.Execute()
}

func hookInterruptHandler () {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("\r- Ctrl+C pressed in Termial\n")
		action.Close()
		os.Exit(0)
	}()
}
