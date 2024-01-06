package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"github.com/burlingtonbertie99/mykeys-ext/auth/fido2"


func main() {
	if len(os.Args) < 2 {
		log.Fatal("specify fido2 library")
	}

	server, err := OpenPlugin(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	req := &DevicesRequest{}
	resp, err := server.Devices(context.TODO(), req)
	if err != nil {
		log.Fatal(err)
	}
	printResponse(resp)
}

func printResponse(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
