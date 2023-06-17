package main

import (
	"fmt"
	"os"
)

func main() {
	var config Config
	if err := ReadConfig("configs/config_example.yml", &config); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("%#v", config)
}
