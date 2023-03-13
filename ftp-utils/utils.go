package ftp_utils

import (
	"fmt"
	"regexp"
)

func IsDataCommand(command string) bool {
	/*
		Returns true if a given command requires a data transfer and is formatted correctly.
	*/
	dataPattern := `^(RETR|STOR) ([a-zA-Z0-9\-_]+)(\.[a-z]+)?$`

	if command == "LIST" {
		return true
	} else if matched, err := regexp.MatchString(dataPattern, command); err == nil && matched {
		return true
	} else if command == "QUIT" {
		return false
	}
	fmt.Println(fmt.Sprintf("Error: Command '%s' requires an argument specifying a filename", command))
	return false
}
