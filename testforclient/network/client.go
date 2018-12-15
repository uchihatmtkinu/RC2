package network

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"encoding/gob"
	"io/ioutil"

	"github.com/uchihatmtkinu/RC/shard"
)

//address
type addr struct {
	AddrList []string
}

//command -> byte
func commandToBytes(command string) []byte {
	var bytees [commandLength]byte

	for i, c := range command {
		bytees[i] = byte(c)
	}

	return bytees[:]
}

//byte -> command
func bytesToCommand(bytees []byte) string {
	var command []byte

	for _, b := range bytees {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

//send data to addr
func sendData(addr string, data []byte) {
	for true {
		conn, err := net.Dial(protocol, addr)
		if err != nil {
			fmt.Println(time.Now(), addr, "is not available")
			return
		}
		defer conn.Close()

		_, err = io.Copy(conn, bytes.NewReader(data))
		if err == nil {
			//log.Panic(err)
			break
		}
		conn.Close()
	}
}

// handle connection
func handleConnection(conn net.Conn, requestChannel chan []byte) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()
	requestChannel <- request

}

//StartServer start a server
func StartServer(ID int) {

	//ln, err := net.Listen(protocol, shard.MyMenShard.Address)
	fmt.Println(bindAddress)
	ln, err := net.Listen(protocol, bindAddress)
	fmt.Println("My IP+Port: ", shard.MyMenShard.Address)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	requestChannel := make(chan []byte, bufferSize)
	flag := true
	IntialReadyCh <- flag
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
		//fmt.Println(time.Now(), "Received", command, "command")

		switch command {
		case "shutDown":
			go HandleShutDown()
		case "requestTxB":
			go HandleRequestTxB(request)
		case "Tx":
			go HandleAndSendTx(request)
		case "TxM":
			go HandleTotalTx(request)
		case "TxMM":
			go HandleTxMM(request)
		case "TxMMRec":
			go HandleTxMMRec(request)
		case "RequestTxMM":
			go HandleRequestTxMM(request)
		case "TxList":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleTxList(request)
		case "TxDec":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleTxDecLeader(request)
		case "TxDecSet":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleAndSentTxDecSet(request)
		/*case "RollRequest":
			go HandleRollingMessage(request)
		case "VTD":
			go HandleVirtualTD(request)
		case "VTDS":
			go HandleVirtualTDS(request)
		case "TxBR":
			go HandleTxBlockAfterRolling(request)
		*/
		case "TxDecSetM":
			//fmt.Printf("%d Received %s command\n", ID, command)
			if shard.GlobalGroupMems[CacheDbRef.ID].Role == 0 {
				go HandleTxDecSetLeader(request)
			} else {
				go HandleTxDecSet(request, 0)
			}
		case "TxDecRev":
			HandleTxDecRev(request)
		case "TxB":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleTxBlock(request)
		//case "FinalTxB":
		//fmt.Printf("%d Received %s command\n", ID, command)
		//go HandleFinalTxBlock(request)
		//case "StartTxB":
		//go HandleStartTxBlock(request)
		//shard
		case "shardReady":
			go HandleShardReady(request)
		case "reqLeaReady":
			go HandleRequestShardLeaderReady(request)
		case "readyAnnoun":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleShardReady(request)
		case "leaderReady":
			//fmt.Printf("%d Received %s command\n", ID, command)
			go HandleLeaderReady(request)
		//rep pow
		case "RepPowAnnou":
			go HandleRepPowRx(request)
		case "RepBlock":
			go HandleRepBlockRx(request)
		case "RequestRep":
			go HandleRequestRepBlockRx(request)

		//cosi protocol
		case "cosiAnnoun":
			go HandleCoSiAnnounce(request)

		case "cosiChallen":
			if CoSiFlag {
				go HandleCoSiChallenge(request)
			}
		case "cosiSig":
			if CoSiFlag {
				go HandleCoSiSig(request)
			}
		case "reqCosiSig":
			go HandleReqCosiSig(request)
		case "cosiCommit":
			if CoSiFlag {
				go HandleCoSiCommit(request)
			}
		case "cosiRespon":
			if CoSiFlag {
				go HandleCoSiResponse(request)
			}
		//sync
		case "requestSync":
			go HandleRequestSync(request)
		case "syncNReady":
			if SyncFlag {
				go HandleSyncNotReady(request)
			}
		case "syncSB":
			if SyncFlag {
				go HandleSyncSBMessage(request)
			}
		case "syncTB":
			if SyncFlag {
				go HandleSyncTBMessage(request)
			}
		case "GossipFirSend":
			go HandleGossipFirSend(request)
		case "GossipFirRev":
			go HandleGossipFirRev(request)
		case "GossipSecSend":
			go HandleGossipSecSend(request)
		case "GossipSecRev":
			go HandleGossipSecRev(request)
		case "QTL":
			go HandleQTL(request)
		case "QTDS":
			go HandleQTDS(request)
		case "QTB":
			go HandleQTB(request)
		default:
			fmt.Println("Unknown command!")
		}

	}
}

//HandleShutDown is to shut down the system
func HandleShutDown() {
	os.Exit(0)
}

//encode
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
