package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

func clientStart() {
	fmt.Println("An welcher Adresse wartet der Host auf eine TCP Verbindung?")
	var address string
	_, err := fmt.Scanln(&address)
	if err != nil {
		fmt.Println("Fehler beim Lesen des Ports:", err)
		os.Exit(1)
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            x509.NewCertPool(),
	}

	// Verbindet sich zum Server
	conn, err := tls.Dial("tcp", address, config)
	if err != nil {
		fmt.Println("Fehler beim Verbinden zum Server:", err)
		os.Exit(1)
	}
	defer func(conn *tls.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Fehler beim schließen der Verbindung:", err)
		}
	}(conn)

	fmt.Println("Verbindung zum Server erfolgreich hergestellt...")

	_, err = conn.Write([]byte("Client Hello\n"))
	if err != nil {
		fmt.Println("Fehler beim begrüßen des Servers:", err)
	}

	reader := bufio.NewReader(conn)

	message, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Fehler beim Lesen von der Verbindung:", err)
		return
	}

	if message != "Server Hello, session key request\n" {
		fmt.Println("Protokollfehler: Kein Server Hello empfangen, verbindung abgebrochen.")
		return
	}

	fmt.Println("Welcher Session möchtest du beitreten?")
	var sessionKey string
	_, err = fmt.Scanln(&sessionKey)
	if err != nil {
		fmt.Println("Fehler beim Lesen des Session Schlüssels:", err)
		return
	}

	_, err = conn.Write([]byte(sessionKey + "\n"))
	if err != nil {
		fmt.Println("Fehler beim senden des Session Keys:", err)
	}

	sessionKey, err = reader.ReadString('\n')
	if err != nil {
		fmt.Println("Fehler beim Lesen von der Verbindung:", err)
		return
	}
	sessionKey = strings.TrimRight(sessionKey, "\n")
	sessionID, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Fehler beim Lesen von der Verbindung:", err)
		return
	}
	sessionID = strings.TrimRight(sessionID, "\n")
	nextUDP, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Fehler beim Lesen von der Verbindung:", err)
		return
	}
	nextUDP = strings.TrimRight(nextUDP, "\n")
	sessionSecret := make([]byte, 32)
	n, err := reader.Read(sessionSecret)
	if err != nil || n != 32 {
		fmt.Println("Fehler beim Lesen von der Verbindung:", err)
		return
	}

	fmt.Println("Session Key:", sessionKey)
	fmt.Println("Session ID:", sessionID)
	fmt.Println("Session Secret:", sessionSecret)
	fmt.Println("Öffne UDP Verbindung:", nextUDP)

	err = conn.Close()
	if err != nil {
		fmt.Println("Fehler beim schließen der Verbindung:", err)
	}

	startChat(sessionKey, sessionID, nextUDP, sessionSecret)
}
