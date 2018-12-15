package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/uchihatmtkinu/RC/rccache"

	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func main() {
	//network.SendTxMessage(gVar.MyAddress, "LogInfo", []byte("Test"))
	//arg, err := strconv.Atoi(os.Args[1])
	/*if err != nil {
		log.Panic(err)
		os.Exit(1)
	}*/
	fmt.Println("Get the local ip from", os.Args[1])
	file, err := os.Open(os.Args[1])
	initType, initErr := strconv.Atoi(os.Args[3])
	if initErr != nil {
		log.Panic(initErr)
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	fileSize := fileinfo.Size()
	buffer := make([]byte, fileSize)

	_, err = file.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Local Address:", string(buffer))

	ID := 0
	totalepoch := 1
	network.IntilizeProcess(string(buffer), &ID, os.Args[2], initType)

	go network.StartServer(ID)
	<-network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)
	numCnt := int(gVar.ShardCnt * gVar.ShardSize)
	tmptx := make([]basic.Transaction, gVar.NumOfTxForTest*gVar.NumTxListPerEpoch)
	//cnt := 0

	time.Sleep(time.Second * 20)
	timestart := time.Now()
	fmt.Println(time.Now(), "test begin")
	//network.StopChan = make(chan os.Signal, 1)
	for k := 1; k <= totalepoch; k++ {
		//test shard
		fmt.Println("Current time: ", time.Now())
		network.ShardProcess()

		rand.Seed(int64(network.CacheDbRef.ID*3000) + time.Now().Unix()%3000)
		for l := 0; l < len(tmptx); l++ {
			i := rand.Int() % numCnt
			for true {
				if basic.ShardIndex(shard.GlobalGroupMems[i].RealAccount.AddrReal) == network.CacheDbRef.ShardNum {
					break
				}
				i = rand.Int() % numCnt
			}
			j := rand.Int() % numCnt
			k := uint32(1)
			tmptx[l] = *rccache.GenerateTx(i, j, k, rand.Int63(), network.CacheDbRef.ID+uint32(l*2000))
			//fmt.Println(base58.Encode(tmptx[l].Hash[:]))
		}

		gVar.T1 = time.Now()
		fmt.Println("This time", time.Now())
		minute := 20
		now := time.Now()
		next := now.Add(0)
		next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), minute, 0, 0, next.Location())
		tt := time.NewTimer(next.Sub(now))
		<-tt.C
		go network.RepGossipLoop(&shard.GlobalGroupMems, minute+1)
		if shard.MyMenShard.Role == shard.RoleLeader {
			fmt.Println("This is a Leader")
			go network.TxGeneralLoop()
			//go network.SendLoopLeader(&tmptx)
			//go network.HandleTxLeader()
		} else {
			go network.MinerLoop()
			//if shard.MyMenShard.InShardId < 50 {
			//go network.SendLoopMiner(&tmptx)
			//}
			//go network.HandleTx()
		}
		//test rep
		//go network.RepProcessLoop(&shard.GlobalGroupMems)
		//<-network.RepFinishChan[gVar.NumberRepPerEpoch-1]
		//test cosi
		/*if network.CacheDbRef.Leader == network.CacheDbRef.ID {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}*/

		//test sync
		//network.SyncProcess(&shard.GlobalGroupMems)
		fmt.Println(time.Now(), "Epoch", k, "finished")
	}
	fmt.Println("Time: ", time.Since(timestart), "TPS:", float64(uint32(totalepoch)*(1+gVar.NumTxListPerEpoch*(gVar.ShardSize-1))*gVar.NumOfTxForTest)/time.Since(timestart).Seconds())
	fmt.Println(network.CacheDbRef.ID, ": All finished")
	if network.CacheDbRef.ID == 0 {
		tmpStr := fmt.Sprint("All finished")
		network.SendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
	}
	time.Sleep(20 * time.Second)

}
