package main

import (
	"UDPChatProgram/crypt"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func startChat(key, id, host string, secret []byte) {
	// Server-Adresse und Port definieren
	fmt.Println("Stelle Verbindung her zu", host)
	serverAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}

	// Verbindung zum Server herstellen
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		os.Exit(1)
	}

	go receive(conn, secret)

	message := pack(key, id, "Neuer Nutzer beigetreten\n", secret)

	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("Fehler beim Anmelden", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		var input string
		if scanner.Scan() {
			input = scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Fehler beim lesen von der Konsole:", err)
			continue
		}

		message := pack(key, id, input+"\n", secret)

		_, err = conn.Write(message)
		if err != nil {
			fmt.Println("Fehler beim versenden der Nachricht", err)
			continue
		}
	}
}

func pack(key, id, msg string, secret []byte) []byte {
	message := []byte(key + id)

	buf := new(bytes.Buffer)
	var length uint32 = uint32(len(msg))
	err := binary.Write(buf, binary.LittleEndian, length)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	message = append(message, buf.Bytes()...)
	encrypted, err := crypt.EncryptAESGCM(secret, []byte(msg))
	if err != nil {
		fmt.Println("Fehler beim Verschlüsseln der Nachricht")
		return nil
	}
	message = append(message, encrypted...)

	return message
}

func receive(conn *net.UDPConn, secret []byte) {
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Fehler beim Lesen von neuen Nachrichten.")
			continue
		}
		message := buffer[:n]

		decrypted, err := crypt.DecryptAESGCM(secret, message)
		if err != nil {
			fmt.Println("Fehler beim Entschlüsseln der Nachricht", err)
			continue
		}
		fmt.Print(string(decrypted))
	}
}
