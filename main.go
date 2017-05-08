package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {

	cli := NewCLI()
	cli.Args = os.Args[1:]

	err := cli.Run()

	if err != nil {
		fmt.Println(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	text, _ := reader.ReadString('\n')
	fmt.Println(text)
	os.Exit(0)
}
