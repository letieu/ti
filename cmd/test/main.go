package main

import (
	"fmt"

	"github.com/letieu/ti/internal/cli"
)

func main() {
	manager, err := cli.NewAuthManager()
	if err != nil {
		fmt.Printf("Err %v \n", err)
	}

	ne := manager.Login("antigravity")
	fmt.Printf("%v a", ne)
}
