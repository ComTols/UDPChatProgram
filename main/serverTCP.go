package main

import (
	"UDPChatProgram/crypt"
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"net"
	"os"
	"strings"
)

func serverStartTCP() {
	// Lädt Zertifikat und Schlüssel
	cert, err := tls.LoadX509KeyPair("crypt/cert.pem", "crypt/key.pem")
	if err != nil {
		fmt.Println("TCP: Fehler beim Laden des Zertifikats:", err)
		os.Exit(1)
	}

	// Konfiguriert TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Deaktiviert die Verifizierung für diese Beispielanwendung
		InsecureSkipVerify: true,
	}

	listen, err := tls.Listen("tcp", "localhost:"+portTCP, config)
	if err != nil {
		fmt.Println("TCP: Fehler beim Starten des Servers:", err)
		os.Exit(1)
	}
	defer func(listen net.Listener) {
		err := listen.Close()
		if err != nil {
			fmt.Println("TCP: Fehler beim schließen der Verbindung:", err)
		}
	}(listen)

	fmt.Printf("TCP: Warten auf eingehende TCP Verbindung an Port %s...\n", portTCP)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("TCP: Fehler beim Akzeptieren der Verbindung:", err)
			continue
		}

		// Handhabung der Verbindung in einer neuen Goroutine
		go handleTCPConnection(conn)
	}
}

func handleTCPConnection(conn net.Conn) {
	fmt.Println("TCP: Verbindung hergestellt:", conn.RemoteAddr())
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("TCP: Fehler beim schließen der Verbindung:", err)
		}
	}(conn)

	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("TCP: Fehler beim Lesen von der Verbindung:", err)
		return
	}

	if message != "Client Hello\n" {
		fmt.Println("TCP: Protokollfehler: Kein Client Hello empfangen, verbindung abgebrochen.")
		return
	}

	fmt.Println("TCP: Verbindung akzeptiert, frage Session Schlüssel...")

	_, err = conn.Write([]byte("Server Hello, session key request\n"))
	if err != nil {
		fmt.Println("TCP: Fehler beim Senden der Antwort:", err)
		return
	}

	sessionKey, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("TCP: Fehler beim lesen des Session Keys:", err)
		return
	}
	sessionKey = strings.TrimRight(sessionKey, "\n")

	fmt.Println("TCP: Angefragter Session Schlüssel:", sessionKey)

	connectionUuid := uuid.New().String()
	connectionSecret, err := crypt.GenerateKey()
	if err != nil {
		fmt.Println("TCP: Fehler beim erstellen eines Secrets:", err)
		return
	}
	var mySession *Session
	if sessionKey == "Neu" {
		fmt.Println("TCP: Neue Session wurde angefordert...")
		newSession := Session{
			clients: []Connection{Connection{id: connectionUuid, secret: connectionSecret}},
			key:     generateSessionKey(),
		}

		for sessionKeyExists(newSession.key) {
			newSession.key = generateSessionKey()
		}

		sessions = append(sessions, newSession)
		mySession = &newSession
		fmt.Println("TCP: Neue Session mit Schlüssel", mySession.key)
	} else {
		fmt.Println("TCP: Suche Session mit angefragtem Schlüssel...")
		founded := false
		for i, session := range sessions {
			if session.key == sessionKey {
				sessions[i].clients = append(session.clients, Connection{id: connectionUuid, secret: connectionSecret})
				founded = true
				mySession = &session
				break
			}
		}
		if !founded {
			fmt.Println("TCP: Der gesuchte Session Key wurde nicht gefunden.")
			return
		}
		fmt.Println("TCP: Session gefunden:", mySession.key)
	}

	_, err = conn.Write(append(append([]byte(mySession.key+"\n"+connectionUuid+"\n"+"localhost:"+portUDP+"\n"), connectionSecret...), []byte("\n")...))
	if err != nil {
		fmt.Println("TCP: Fehler beim Senden des Session Keys:", err)
		return
	}
}
