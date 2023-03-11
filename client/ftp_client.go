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
		command      string
		dataCommands = []string{"LIST", "RETR", "STOR"}
		pattern      = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$`
		buffer       = make([]byte, 1024)
	)

	//establish connection
	for {
		fmt.Println("Connect to a server:")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			command = scanner.Text()
		}

		if matched, err := regexp.MatchString(pattern, command); err == nil && matched {
			splitCommand := strings.Split(command, " ")
			hostAndPort := strings.Split(splitCommand[1], ":")
			host := hostAndPort[0]
			port := hostAndPort[1]

			//Control Connection
			control, err := establishConnection("Control", host, port)
			if err != nil {
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
				if isDataCommand(dataCommands, command) {
					fmt.Println("Data Port Running on " + SERVER_HOST + ":" + SERVER_PORT)
					server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)

					if err != nil {
						fmt.Println("Error listening:", err.Error())
						os.Exit(1)
					}
					_, _ = control.Write([]byte(command))

					connection, err := server.Accept()
					if err != nil {
						fmt.Println("Error accepting: ", err.Error())
						os.Exit(1)
					}

					dataLength, err := connection.Read(buffer)
					if err != nil {
						fmt.Println("Error reading:", err.Error())
						return
					}
					dataToString := string(buffer[:dataLength])
					fmt.Println(dataToString)

					fmt.Println("Data Port Closing")
					connection.Close()
					server.Close()
				} else if command == "QUIT" {
					control.Write([]byte("QUIT"))
					break
				} else {
					fmt.Println("Invalid command. Try again")
				}
			}
			err = control.Close()
			if err != nil {
				return
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

func isDataCommand(commands []string, command string) bool {
	for _, value := range commands {
		if value == command {
			return true
		}
	}
	return false
}
