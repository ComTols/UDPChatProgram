package main

import (
	"UDPChatProgram/crypt"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func startServerUDP() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:"+portUDP)
	if err != nil {
		fmt.Println("UDP: Fehler beim Auflösen der Adresse:", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("UDP: Fehler beim Erstellen des Sockets:", err)
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("UDP: Fehler beim schließen der Verbindung.")
		}
	}(conn)

	fmt.Printf("UDP: Warten auf eingehende UDP Verbindung an Port %s...\n", portUDP)

	for {
		handleConnection(conn)
	}
}

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println("UDP: Fehler beim Lesen von der Verbindung:", err)
		return
	}
	fmt.Printf("UDP: Empfangene Daten von %s\n", addr)

	sessionKey, connectionID, message, err := parsePacket(conn, buffer[:n])
	if err != nil {
		fmt.Println("UDP: Fehler beim Parsen des Pakets:", err)
		return
	}

	var mySession *Session
	var myConnection *Connection
	founded := false
	for _, session := range sessions {
		if session.key == sessionKey {
			mySession = &session
			for i, connection := range session.clients {
				if connection.id == connectionID {
					session.clients[i].conn = conn
					session.clients[i].addr = addr
					myConnection = &connection
					founded = true
					break
				}
			}
			break
		}
	}

	if !founded || mySession == nil || myConnection == nil {
		fmt.Println("UDP: Session wurde nicht gefunden.", founded, mySession, myConnection)
		return
	}

	decrypted, err := crypt.DecryptAESGCM(myConnection.secret, message)
	if err != nil {
		fmt.Println("UDP: Fehler beim entschlüsseln der Nachricht", err)
		return
	}
	fmt.Printf("UDP: SessionKey: %s, ConnectionID: %s, Nachricht: %s\n", sessionKey, connectionID, strings.TrimRight(string(decrypted), "\n"))

	for _, connection := range mySession.clients {
		if connection.conn == nil && connection.id != connectionID {
			fmt.Printf("UDP: Client %s hat noch keine Verbindung hergestellt.\n", connection.id)
			continue
		}
		if connection.id != connectionID {
			encrypted, err := crypt.EncryptAESGCM(connection.secret, decrypted)
			if err != nil {
				fmt.Println("UDP: Fehler beim Verschlüsseln der Nachricht", err)
				continue
			}
			_, err = connection.conn.WriteToUDP(encrypted, connection.addr)
			if err != nil {
				fmt.Println("UDP: Fehler beim senden der Nachricht", err)
			}
		}
	}
}

func parsePacket(conn *net.UDPConn, initialPacket []byte) (string, string, []byte, error) {
	if len(initialPacket) < 46 {
		return "", "", nil, fmt.Errorf("paket zu kurz")
	}

	sessionKey := string(initialPacket[:6])
	connectionID := string(initialPacket[6:42])

	var msgLength int32
	msgLengthBytes := initialPacket[42:46]
	buf := bytes.NewReader(msgLengthBytes)
	err := binary.Read(buf, binary.LittleEndian, &msgLength)
	if err != nil {
		return "", "", nil, fmt.Errorf("fehler beim Lesen der Nachrichtenlänge: %v", err)
	}

	remainingBytes := int(msgLength) - len(initialPacket[46:])
	message := initialPacket[46:]

	if remainingBytes > 0 {
		// If there are still bytes to read, continue reading from the connection
		tempBuffer := make([]byte, remainingBytes)
		n, _, err := conn.ReadFromUDP(tempBuffer)
		if err != nil {
			return "", "", nil, fmt.Errorf("fehler beim Lesen der verbleibenden Nachricht: %v", err)
		}
		message = append(message, tempBuffer[:n]...)
	}

	return sessionKey, connectionID, message, nil
}
