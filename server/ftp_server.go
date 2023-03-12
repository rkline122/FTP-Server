package main

import (
	"fmt"
	"net"
	"os"
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
		buffer       = make([]byte, 1024)
		dataCommands = []string{"LIST", "RETR", "STOR"}
	)

	for {
		// Reads and deconstructs client message
		messageLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("[Control] Error reading:", err.Error())
			return
		}
		command := string(buffer[:messageLen])

		if isDataCommand(dataCommands, command) {
			var data net.Conn
			for {
				data, err = establishConnection("Data", SERVER_HOST, "8000")
				if err != nil {
					continue
				}
				break
			}
			_, err := data.Write([]byte(fmt.Sprintf("<insert %s data here>", command)))
			if err != nil {
				fmt.Println("[Data] Error writing:", err.Error())
				return
			}
			err = data.Close()
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

func isDataCommand(commands []string, command string) bool {
	for _, value := range commands {
		if value == command {
			return true
		}
	}
	return false
}
