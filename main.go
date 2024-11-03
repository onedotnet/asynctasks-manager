package main

import "github.com/onedotnet/asynctasks/cmd"

func main() {
	// Code
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
