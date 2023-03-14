/*
Project 2: FTP Server
By Ryan Kline
		---
CIS 457 - Data Communications
Winter 2023
*/

package main

import (
	"bufio"
	"fmt"
	"io"
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
	/*
		Prompts user to connect to the server using the command 'CONNECT <server name/IP address>:<port>'.

		Upon successful connection, the user is able to send commands to the server. If a command requires a
		data transfer, a server is started on the client to act as the data connection. Once the FTP server has
		been connected to the data line, the handleDataTransfer() function is called and runs the appropriate logic
		based on the command sent. When the transfer is complete, the data connection is closed and the user is
		prompted to send another command. This loop continues until the client sends the "QUIT" command.
	*/
	var (
		command        string
		connectPattern = `^CONNECT ([a-zA-Z0-9\-\.]+:[0-9]+)$`
	)

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

			// Control Connection
			control, err := net.Dial(SERVER_TYPE, host+":"+port)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(fmt.Sprintf("[Control] Connected to %s:%s", host, port))
			}

			// Send Host/Port info for data connection
			_, err = control.Write([]byte(SERVER_HOST + ":" + SERVER_PORT))
			if err != nil {
				fmt.Println("Unable to write to server:", err.Error())
				return
			}

			// Interact with the server via commands
			for {
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Println("Enter a command:")

				if scanner.Scan() {
					command = scanner.Text()
				}

				if command != "QUIT" {
					if isValidCommand(command) {
						fmt.Println("[Data] Port Running on " + SERVER_HOST + ":" + SERVER_PORT)
						server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
						if err != nil {
							fmt.Println("[Data] Error listening:", err.Error())
							return
						}

						_, err = control.Write([]byte(command))
						if err != nil {
							fmt.Println("[Control] Error writing:", err.Error())
							return
						}

						dataConnection, err := server.Accept()
						if err != nil {
							fmt.Println("[Data] Error accepting client:", err.Error())
							return
						}

						err = handleDataTransfer(command, dataConnection)
						if err != nil {
							fmt.Println("[Data] Error in data transfer:", err.Error())
							return
						}

						fmt.Println("[Data] Port Closing")
						err = dataConnection.Close()
						if err != nil {
							fmt.Println("[Data] Error closing dataConnection to client:", err.Error())
							return
						}

						err = server.Close()
						if err != nil {
							fmt.Println("[Data] Error closing server:", err.Error())
							return
						}
					}
				} else if command == "QUIT" {
					_, err := control.Write([]byte(command))
					if err != nil {
						fmt.Println("[Control] Error writing:", err.Error())
						return
					}
					break
				} else {
					fmt.Println("Invalid command. Try again")
				}
			}
			err = control.Close()
			if err != nil {
				fmt.Println("[Control] Error closing connection to server:", err.Error())
				return
			}
		} else if command == "exit" {
			os.Exit(0)
		} else {
			fmt.Println("Invalid Command. Use `CONNECT <server name/IP address> <server port>` to connect to " +
				"a server")
		}
	}
}

func isValidCommand(command string) bool {
	/*
		Returns true if a given command is valid.
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
	buffer := make([]byte, 1024)

	if instruction == "LIST" {
		//	Read from data, print contents to terminal
		dataLength, err := dataConnection.Read(buffer)
		if err != nil {
			fmt.Println("[Data] Error reading from client:", err.Error())
			return err
		}
		dataToString := string(buffer[:dataLength])
		fmt.Println(dataToString)
	} else {
		splitInstruction := strings.Split(instruction, " ")
		command := splitInstruction[0]
		filename := splitInstruction[1]

		if command == "STOR" {
			//	Send file to the server
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
		} else if command == "RETR" {
			//	Retrieve file from the server
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
		}
	}
	return nil
}
