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

	var (
		buffer = make([]byte, 1024)
	)

	for {
		// Reads and deconstructs client message
		messageLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[Control] Error reading:", err.Error())
			return
		}
		command := string(buffer[:messageLen])

		if isDataCommand(command) {
			var dataConnection net.Conn
			for {
				dataConnection, err = establishConnection("Data", SERVER_HOST, "8000")
				if err != nil {
					continue
				}
				break
			}

			// --------- Handle Instructions Here-----------
			err := handleInstruction(command, dataConnection)
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

func establishConnection(connectionType, host string, port string) (net.Conn, error) {
	connection, err := net.Dial(SERVER_TYPE, host+":"+port)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(fmt.Sprintf("[%s] Connected to %s:%s", connectionType, host, port))
	}
	return connection, err
}

func isDataCommand(command string) bool {
	argsPattern := `^(RETR|STOR) ([a-zA-Z0-9\-_]+)(\.[a-z]+)?$`

	if command == "LIST" {
		return true
	} else if matched, err := regexp.MatchString(argsPattern, command); err == nil && matched {
		return true
	}
	fmt.Println(fmt.Sprintf("Error: Command '%s' requires an arguement specifying a filename", command))

	return false
}

func handleInstruction(instruction string, dataConnection net.Conn) error {
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
