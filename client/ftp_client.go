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
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8636"
	SERVER_TYPE = "tcp"
)

func main() {
	var (
		command string
		msg     string
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
			connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
			if err != nil {
				panic(err)
			} else {
				fmt.Println(fmt.Sprintf("Connected to %s:%s", SERVER_HOST, SERVER_PORT))
			}

			///send some data
			for {
				fmt.Println("Write a message to send to the server:")
				fmt.Scanln(&msg)

				if msg != "QUIT" {
					_, _ = connection.Write([]byte(msg))
					buffer := make([]byte, 1024)
					mLen, err := connection.Read(buffer)
					if err != nil {
						fmt.Println("Error reading:", err.Error())
					}
					fmt.Println("Received: ", string(buffer[:mLen]))
				} else {
					fmt.Println(fmt.Sprintf("Ending connection with %s:%s", SERVER_HOST, SERVER_PORT))
					break
				}
			}
			defer connection.Close()
		} else if command == "QUIT" {
			break
		} else {
			fmt.Println("Command does not match pattern")
		}
	}

}
