// The client program presents a command line interface that allows a user to:
//
//    Connect to a server
//    List files located at the server.
//    Get (retrieve) a file from the server.
//    Send (put) a file from the client to the server.
//    Terminate the connection to the server.

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8000"
	SERVER_TYPE = "tcp"
)

func main() {
	var (
		command        string
		connectPattern = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$` // CONNECT localhost:8636
	)

	//establish connection
	for {
		fmt.Println("Connect to a server:")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			command = scanner.Text()
		}

		if matched, err := regexp.MatchString(connectPattern, command); err == nil && matched {
			splitCommand := strings.Split(command, " ")
			hostAndPort := strings.Split(splitCommand[1], ":")
			host := hostAndPort[0]
			port := hostAndPort[1]

			//Control Connection
			control, err := establishConnection("Control", host, port)
			if err != nil {
				// Error message prints in establishConnection function
				continue
			}

			// Interact with the server via commands
			for {
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("Enter a command:")

				if scanner.Scan() {
					command = scanner.Text()
				}

				// If client expects data, open the data port
				if isDataCommand(command) {
					fmt.Println("[Data] Port Running on " + SERVER_HOST + ":" + SERVER_PORT)
					server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
					if err != nil {
						fmt.Println("[Data] Error listening:", err.Error())
						continue
					}

					_, err = control.Write([]byte(command))
					if err != nil {
						fmt.Println("[Control] Error writing:", err.Error())
						continue
					}

					connection, err := server.Accept()
					if err != nil {
						fmt.Println("[Data] Error accepting client:", err.Error())
						continue
					}

					//dataLength, err := connection.Read(buffer)
					//if err != nil {
					//	fmt.Println("[Data] Error reading from client:", err.Error())
					//	continue
					//}
					//dataToString := string(buffer[:dataLength])
					//fmt.Println(dataToString)

					fmt.Println("[Data] Port Closing")
					err = connection.Close()
					if err != nil {
						fmt.Println("[Data] Error closing connection to client:", err.Error())
						continue
					}
					err = server.Close()
					if err != nil {
						fmt.Println("[Data] Error closing server:", err.Error())
						continue
					}
				} else if command == "QUIT" {
					_, err := control.Write([]byte("QUIT"))
					if err != nil {
						fmt.Println("[Control] Error writing:", err.Error())
						os.Exit(1)
					}
					break
				} else {
					fmt.Println("Invalid command. Try again")
				}
			}
			err = control.Close()
			if err != nil {
				fmt.Println("[Control] Error closing connection to server:", err.Error())
				os.Exit(1)
			}
		} else if command == "QUIT" {
			break
		} else {
			fmt.Println("Invalid Command. Use `CONNECT <server name/IP address> <server port>` to connect to " +
				"a server")
		}
	}
}

func establishConnection(connectionType, host, port string) (net.Conn, error) {
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

func handleDataTransfer(instruction string, dataConnection net.Conn) error {
	buffer := make([]byte, 1024)

	if instruction == "LIST" {
		//	Read from data, print contents to terminal
	} else {
		splitInstruction := strings.Split(instruction, " ")
		command := splitInstruction[0]
		filename := splitInstruction[1]

		if command == "STOR" {
			//	Send file to the server

		} else if command == "RETR" {
			//	Retrieve file from the server

		}
	}
	return nil
}