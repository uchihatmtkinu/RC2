package main

import (
	"fmt"
	"time"

	"github.com/uchihatmtkinu/RC/Reputation"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/shard"
	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func main() {
	ID := 3
	totalepoch := 2
	network.IntilizeProcess(ID)
	fmt.Println("test begin")
	go network.StartServer(ID)
	<-network.IntialReadyCh
	close(network.IntialReadyCh)

	fmt.Println("MyGloablID: ", network.MyGlobalID)
	for k := 1; k <= totalepoch; k++ {
		//test shard
		network.ShardProcess()

		if k == 1 {
			go network.TxGeneralLoop()
		}
		//test rep
		network.RepProcess(&shard.GlobalGroupMems)
		Reputation.CurrentRepBlock.Mu.RLock()
		Reputation.CurrentRepBlock.Block.Print()
		Reputation.CurrentRepBlock.Mu.RUnlock()
		fmt.Println(network.CacheDbRef.ID, "goes into Sync procedure")
		//test cosi
		if shard.MyMenShard.Role == shard.RoleLeader {
			network.LeaderCosiProcess(&shard.GlobalGroupMems)
		} else {
			network.MemberCosiProcess(&shard.GlobalGroupMems)
		}

		//test sync
		network.SyncProcess(&shard.GlobalGroupMems)
		time.Sleep(10 * time.Second)

		Reputation.CurrentSyncBlock.Mu.RLock()
		Reputation.CurrentSyncBlock.Block.Print()
		Reputation.CurrentSyncBlock.Mu.RUnlock()
		network.CacheDbRef.Mu.Lock()
		fmt.Println("FB from", network.CacheDbRef.ID)
		for i := uint32(0); i < gVar.ShardCnt; i++ {
			network.CacheDbRef.FB[i].Print()
		}
		network.CacheDbRef.Mu.Unlock()

		for i := 0; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].Print()
		}

		for i := 0; i < int(gVar.ShardSize*gVar.ShardCnt); i++ {
			shard.GlobalGroupMems[i].Print()
			fmt.Println()
		}
	}

	fmt.Println("All finished")

	time.Sleep(600 * time.Second)
}
