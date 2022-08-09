package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lemon-mint/envaddr"
)

func main() {
	srv := http.Server{}
	ln, err := net.Listen("tcp", envaddr.Get(":8080"))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Printf("Listening on %s\n", ln.Addr())
	go srv.Serve(ln)

	{
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("\nShutting down...")
		err = srv.Shutdown(nil)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		fmt.Println("Done.")
	}
}
