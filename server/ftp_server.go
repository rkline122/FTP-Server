/*
Project 2: FTP Server
By Ryan Kline
		---
CIS 457 - Data Communications
Winter 2023
*/

package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8636"
	SERVER_TYPE = "tcp"
)

func main() {
	/*
	   Starts up server using the host, port, and
	   protocol defined above. Once a client is connected,
	   the processClient() function is ran as a goroutine (new thread)
	*/
	fmt.Println("Server Running...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer func(server net.Listener) {
		err := server.Close()
		if err != nil {
			fmt.Println("Cannot close server:", err.Error())
			os.Exit(1)
		}
	}(server)

	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")

	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go processClient(connection)
	}
}

func processClient(connection net.Conn) {
	/*
		Listens for commands sent from the client. If a command requires a data transfer, the server connects to
		the data line hosted by the client before calling the handleDataTransfer() function that takes appropriate actions
		based on the instruction received. Once the transfer is complete, it closes its end of the data line and waits
		for a new instruction from the client. This process continues until "QUIT" is received from the client.
	*/

	var (
		buffer   = make([]byte, 1024)
		dataHost string
		dataPort string
	)

	// Receives and stores host and port number for data connection
	messageLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("[Control] Error reading:", err.Error())
		return
	}
	bufferToString := string(buffer[:messageLen])
	dataHostAndPort := strings.Split(bufferToString, ":")
	dataHost = dataHostAndPort[0]
	dataPort = dataHostAndPort[1]

	for {
		// Reads and deconstructs client message
		messageLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[Control] Error reading:", err.Error())
			return
		}
		command := string(buffer[:messageLen])

		if command != "QUIT" {
			if isValidCommand(command) {
				dataConnection, err := net.Dial(SERVER_TYPE, dataHost+":"+dataPort)
				if err != nil {
					return
				}
				fmt.Println(fmt.Sprintf("[Data] Connected to %s:%s", dataHost, dataPort))

				err = handleDataTransfer(command, dataConnection)
				if err != nil {
					fmt.Println("[Data] Error executing instruction:", err.Error())
					return
				}

				err = dataConnection.Close()
				if err != nil {
					fmt.Println("[Data] Error closing connection to server:", err.Error())
					return
				}
				fmt.Println("[Data] Connection closed")
			}
		} else if command == "QUIT" {
			err = connection.Close()
			if err != nil {
				fmt.Println("[Control] Error closing connection to client:", err.Error())
				return
			}
			fmt.Println("Connection Ended by Client")
			break
		}
	}
}

func isValidCommand(command string) bool {
	/*
		Returns true if a given command requires a data transfer and is formatted correctly.
	*/
	dataPattern := `^(RETR|STOR) ([a-zA-Z0-9\-_]+)(\.[a-z]+)?$`
	matched, err := regexp.MatchString(dataPattern, command)

	if command == "LIST" || matched && err == nil {
		return true
	}
	fmt.Println("Invalid command or incorrect format. (Make sure to include the filename for STOR and RETR)")
	return false
}

func handleDataTransfer(instruction string, dataConnection net.Conn) error {
	/*
		Executes appropriate actions based on the command passed. Returns any potential errors.
	*/
	if instruction == "LIST" {
		// Build a string that contains all files in the current directory, send to client
		data := ""
		files, err := os.ReadDir(".")

		if err != nil {
			return err
		}

		for _, file := range files {
			data += file.Name() + " "
		}

		_, err = dataConnection.Write([]byte(data))
		if err != nil {
			fmt.Println("[Data] Error writing:", err.Error())
			return err
		}
	} else {
		splitInstruction := strings.Split(instruction, " ")
		command := splitInstruction[0]
		filename := splitInstruction[1]

		if command == "STOR" {
			//	Receive a file from the client
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err.Error())
				return err
			}
			_, err = io.Copy(file, dataConnection)
			if err != nil {
				fmt.Println("Error copying data: ", err.Error())
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println("Error closing file: ", err.Error())
				return err
			}
		} else if command == "RETR" {
			// Send a file to the client
			file, err := os.Open("./" + filename)
			if err != nil {
				fmt.Println(err)
				return err
			}
			_, err = io.Copy(dataConnection, file)
			if err != nil {
				fmt.Println(err)
				return err
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
	}
	return nil
}
