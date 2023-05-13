package main

import (
	"flag"
	"log"
	"net"
	"strings"

	"github.com/armon/go-socks5"
)

var (
	listenAddr = flag.String("listen-addr", "127.0.0.1:1080", "proxy server listen address")
	user       = flag.String("user", "123", "proxy authentication username")
	pass       = flag.String("pass", "321", "proxy authentication password")
)

func main() {
	flag.Parse()

	// Create a SOCKS5 server with authentication.
	config := &socks5.Config{}
	if *user != "" && *pass != "" {
		creds := socks5.StaticCredentials{
			*user: *pass,
		}
		config.Credentials = creds
	}
	server, err := socks5.New(config)
	if err != nil {
		log.Fatalf("failed to create SOCKS5 server: %v", err)
	}

	// Listen for incoming connections.
	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("failed to listen on %v: %v", *listenAddr, err)
	}
	log.Printf("SOCKS5 proxy server listening on %v", *listenAddr)

	// Accept and handle incoming connections.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept incoming connection: %v", err)
			continue
		}
		go func() {
			defer conn.Close()
			if err := server.ServeConn(conn); err != nil {
				if strings.Contains(err.Error(), "EOF") {
					log.Printf("client closed connection: %v", err)
				} else {
					log.Printf("failed to serve client connection: %v", err)
				}
			}
		}()
	}
}
