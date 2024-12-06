package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var _ = fmt.Fprint

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error In User Input")
		}
		command = strings.TrimSpace(command)
		fmt.Println(command + ": command not found")
	}
}
