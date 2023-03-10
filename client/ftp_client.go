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
	SERVER_TYPE = "tcp"
)

func main() {
	var (
		command string
		pattern = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$`
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

			///send some data
			for {
				fmt.Println("Enter a command:")
				if scanner.Scan() {
					command = scanner.Text()
				}

				//TODO: Do some preprocessing to separate the command type from potential command args
				//TODO: Make create an array of all valid command types

				//Data Connection - established and torn down after each command
				data, err := establishConnection("Data", host, port)
				if err != nil {
					continue
				}

				if command != "QUIT" {
					_, _ = control.Write([]byte(command))
					buffer := make([]byte, 1024)
					mLen, err := control.Read(buffer)
					if err != nil {
						fmt.Println("Error reading:", err.Error())
					}
					fmt.Println("Received: ", string(buffer[:mLen]))
					defer func(data net.Conn) {
						err := data.Close()
						if err != nil {
							fmt.Println(err)
						}
					}(data)
				} else {
					fmt.Println(fmt.Sprintf("Ending connection with %s:%s", host, port))
					defer data.Close()
					break
				}

			}
			defer control.Close()

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
