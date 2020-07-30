package main

import (
	"fmt"
	"os"
)

func writeFile(txt string) {

	if _, err := os.Stat("result.log"); os.IsNotExist(err) {
		fo, err := os.Create("result.log")
		if err != nil {
			fmt.Println("Cannot create file.")
		}
		fo.Close()
	}
	fo, err := os.OpenFile("result.log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	fo.WriteString(txt + "\n")
}
