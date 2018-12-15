package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func main() {
	go network.StartLocalServer()
	PubfileIP, err := os.Open(os.Args[1])
	defer PubfileIP.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var tmp int
	flag := true
	for flag {
		fmt.Scanln(&tmp)
		if tmp == 1 {
			flag = false
		}
	}
	initType, initErr := strconv.Atoi(os.Args[2])
	if initErr != nil {
		log.Panic(initErr)
		os.Exit(1)
	}
	scannerPub := bufio.NewScanner(PubfileIP)
	scannerPub.Split(bufio.ScanWords)
	IPCnt := gVar.ShardCnt * gVar.ShardSize
	if initType != 0 {
		IPCnt = gVar.ShardCnt * gVar.ShardSize / 2
	}
	for i := 0; i < int(IPCnt); i++ {

		scannerPub.Scan()
		IPAddrPub := scannerPub.Text()

		IPAddr2 := IPAddrPub + ":" + strconv.Itoa(3000+i)
		//IPAddr2 := IPAddrPub + ":"
		network.SendTxMessage(IPAddr2, "shutDown", []byte(""))
		if initType != 0 {
			IPAddr2 = IPAddrPub + ":" + strconv.Itoa(3000+i+int(IPCnt))
			network.SendTxMessage(IPAddr2, "shutDown", []byte(""))
		}
	}
	fmt.Println("all shut down")
}
