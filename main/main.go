package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
)

var (
	sessions []Session
	portTCP  string
	portUDP  string
)

type Session struct {
	clients []Connection
	key     string
}

type Connection struct {
	conn   *net.UDPConn
	addr   *net.UDPAddr
	secret []byte
	id     string
}

func main() {
	fmt.Println("Chat hosten? [yes]")
	var host string
	_, err := fmt.Scanln(&host)
	if err != nil {
		fmt.Println("Fehler beim Lesen:", err)
		os.Exit(1)
	}

	if host == "yes" {
		fmt.Println("An welchem Port soll auf eine TCP Verbindung gewartet werden?")
		_, err := fmt.Scanln(&portTCP)
		if err != nil {
			fmt.Println("Fehler beim Lesen des Ports:", err)
			os.Exit(1)
		}

		fmt.Println("An welchem Port soll auf eine UDP Verbindung gewartet werden?")
		_, err = fmt.Scanln(&portUDP)
		if err != nil {
			fmt.Println("Fehler beim Lesen des Ports:", err)
			os.Exit(1)
		}
		go serverStartTCP()
		go startServerUDP()
		select {}
	} else {
		clientStart()
	}
}

func generateSessionKey() string {
	var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func sessionKeyExists(key string) bool {
	for _, session := range sessions {
		if session.key == key {
			return true
		}
	}
	return false
}
