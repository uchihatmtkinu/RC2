package network

import (
	"fmt"
	"log"
	"net"
	"time"
)

//StartLocalServer start a server
func StartLocalServer() {

	//ln, err := net.Listen(protocol, shard.MyMenShard.Address)
	//fmt.Println(bindAddress)
	ln, err := net.Listen(protocol, ":9999")
	//fmt.Println("My IP+Port: ", shard.MyMenShard.Address)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	requestChannel := make(chan []byte, bufferSize)

	fmt.Println("intial ready")
	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, requestChannel)

		request := <-requestChannel
		if len(request) < commandLength {
			continue
		}
		command := bytesToCommand(request[:commandLength])
		if len(request) > commandLength {
			request = request[commandLength:]
		}
		//fmt.Println(time.Now(), ID, "Received", command, "command")

		switch command {

		case "LogInfo":
			fmt.Println(time.Now(), string(request))
		default:
			fmt.Println("Unknown command!")
		}
	}
}
