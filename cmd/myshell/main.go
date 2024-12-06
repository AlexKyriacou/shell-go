package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var _ = fmt.Fprint

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		commandRaw, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error In User Input")
		}
		command := strings.Split(strings.TrimSpace(commandRaw), " ")
		switch command[0] {
		case "exit":
			exit(command[1:])
		case "echo":
			fmt.Println(strings.Join(command[1:], " "))
		default:
			fmt.Println(command[0] + ": command not found")
		}
	}
}

func exit(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command exit. Expected 1, Got", len(args))
	} else {
		exitCode, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid exit code " + args[0])
		}
		os.Exit(exitCode)
	}
}
