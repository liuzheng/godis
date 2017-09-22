package main

import (
	"os"
	"github.com/liuzheng712/godis/godis"
	"fmt"
)

func main() {
	gedis, err := godis.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	gedis.Run()
}
