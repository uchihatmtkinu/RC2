package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/rccache"

	"github.com/uchihatmtkinu/RC/gVar"

	"github.com/uchihatmtkinu/RC/shard"

	"github.com/uchihatmtkinu/RC/basic"
)

// SendTxMessage send reputation block
func SendTxMessage(addr string, command string, message []byte) {
	tmp := make([]byte, len(message))
	copy(tmp, message)
	sendTxMessage(addr, command, tmp)
}

// sendTxMessage send reputation block
func sendTxMessage(addr string, command string, message []byte) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

//TxGeneralLoop is the normall loop of transaction cache
func TxGeneralLoop() {
	rand.Seed(time.Now().Unix() * int64(CacheDbRef.ID))

	fmt.Println(time.Now(), CacheDbRef.ID, "start to process Tx:")
	if CacheDbRef.Now == nil {
		CacheDbRef.NewTxList()
	}
	for i := 0; i < gVar.NumTxListPerEpoch; i++ {
		//<-StartNewTxlist
		time.Sleep(gVar.TxSendInterval * time.Second)
		go TxListProcess()
	}
}

//TxListProcess is the process for txlist
func TxListProcess() {
	CacheDbRef.Mu.Lock()
	CacheDbRef.BuildTDS()
	TLG := CacheDbRef.Now
	fmt.Println(time.Now(), CacheDbRef.ID, "sends a TxList with", TLG.TLS[CacheDbRef.ShardNum].TxCnt, "Txs, Hash:", base58.Encode(TLG.TLS[CacheDbRef.ShardNum].HashID[:]))
	tmpStr := fmt.Sprintln("Shard", CacheDbRef.ShardNum, ":", CacheDbRef.ID, "sends a TxList with", TLG.TLS[CacheDbRef.ShardNum].TxCnt, "Txs Round", TLG.TLS[CacheDbRef.ShardNum].Round)
	sendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))

	//CacheDbRef.TLS[CacheDbRef.ShardNum].Print()
	data1 := new([]byte)
	thisround := TLG.TLS[CacheDbRef.ShardNum].Round
	TLG.TLS[CacheDbRef.ShardNum].Encode(data1)
	CacheDbRef.TLSCache[thisround] = &TLG.TLS[CacheDbRef.ShardNum]
	go SendTxList(*data1, gVar.GossipRound)
	CacheDbRef.NewTxList()
	CacheDbRef.Mu.Unlock()
	cnt := 1
	timeoutflag := true
	for timeoutflag && cnt < int(gVar.ShardSize) {
		select {
		case <-TLChan[thisround-CacheDbRef.PrevHeight]:
			cnt++
			//fmt.Println("Get TxDec of", base58.Encode(TLG.TLS[CacheDbRef.ShardNum].HashID[:]))
		case <-time.After(timeoutTL):
			fmt.Println("TxDecSet is not full", cnt, "in total")
			timeoutflag = false
		}
	}
	fmt.Println("TxDec of Round", thisround, "total txdec: ", uint32(cnt))
	tmpflag := false
	if thisround-CacheDbRef.PrevHeight != 0 {
		<-TBChan[thisround-CacheDbRef.PrevHeight-1]
	}
	CacheDbRef.Mu.Lock()

	fmt.Println(time.Now(), "Leader", CacheDbRef.ID, "ready to send TDS Hash:", base58.Encode(TLG.TLS[CacheDbRef.ShardNum].HashID[:]))
	fmt.Println("TLG TDS length", len(TLG.TDS))
	CacheDbRef.SignTDS(TLG)
	CacheDbRef.ProcessTDS(&TLG.TDS[CacheDbRef.ShardNum], &CacheDbRef.RepCache[thisround-CacheDbRef.PrevHeight])
	fmt.Println(time.Now(), CacheDbRef.ID, "sends a TxDecSet with hash:", base58.Encode(TLG.TDS[CacheDbRef.ShardNum].HashID[:]))
	data2 := new([][]byte)
	*data2 = make([][]byte, gVar.ShardCnt)

	for i := uint32(0); i < gVar.ShardCnt; i++ {
		TLG.TDS[i].Encode(&(*data2)[i])
	}
	CacheDbRef.TDSCache[thisround] = &TLG.TDS[0]
	go SendTxDecSet(*data2, thisround-CacheDbRef.PrevHeight, gVar.GossipRound)
	go TxNormalBlock(thisround - CacheDbRef.PrevHeight)

	CacheDbRef.Release(TLG)

	CacheDbRef.TDSCnt[CacheDbRef.ShardNum]++
	if CacheDbRef.TDSCnt[CacheDbRef.ShardNum] == gVar.NumTxListPerEpoch {
		CacheDbRef.TDSNotReady--
		fmt.Println("Decrease the TDSCnt to", CacheDbRef.TDSNotReady)
		if CacheDbRef.TDSNotReady == 0 {
			tmpflag = true
		}
	}
	CacheDbRef.Mu.Unlock()
	if thisround-CacheDbRef.PrevHeight < gVar.NumTxListPerEpoch-1 {
		TBChan[thisround-CacheDbRef.PrevHeight] <- 1
	}
	if tmpflag {
		StartLastTxBlock <- CurrentEpoch
	}
}

/*
//TxLastBlock is the txlastblock
func TxLastBlock() {
	lastFlag := true
	for lastFlag {
		tmpEpoch := <-StartLastTxBlock
		if tmpEpoch == CurrentEpoch {
			lastFlag = false
		}
	}
	if gVar.ExperimentBadLevel == 0 || !CacheDbRef.Badness {
		CacheDbRef.Mu.Lock()
		CacheDbRef.GenerateTxBlock(1)
		fmt.Println(time.Now(), CacheDbRef.ID, "sends the last TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height", CacheDbRef.TxB.Height)
		tmpStr := fmt.Sprintln("Shard", CacheDbRef.ShardNum, ":", CacheDbRef.ID, "sends last TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height:", CacheDbRef.TxB.Height)
		sendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
		data3 := new([]byte)
		CacheDbRef.TxB.Encode(data3, 0)
		go SendTxBlock(data3, 0, gVar.GossipRound)

		for i := CacheDbRef.TxB.Height - uint32(len(*CacheDbRef.TBCache)) - CacheDbRef.PrevHeight; i < CacheDbRef.TxB.Height-1-CacheDbRef.PrevHeight; i++ {
			//fmt.Println("Rep prepare: Round", i)
			//fmt.Println(CacheDbRef.RepCache[i])
			for j := uint32(0); j < gVar.ShardSize; j++ {
				shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][j]].Rep += CacheDbRef.RepCache[i][j]
			}
		}
		tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
		tmp := make([][32]byte, len(*CacheDbRef.TBCache))
		copy(tmp, *CacheDbRef.TBCache)
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[len(*CacheDbRef.TBCache):]
		CacheDbRef.Mu.Unlock()
		CurrentRepRound++
		fmt.Println(time.Now(), CacheDbRef.ID, "start to make last repBlock, Round:", CurrentRepRound)
		go LeaderCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: false, Hash: tmp, Rep: tmpRep, Round: CurrentRepRound})

		//StopGetTx <- true
		fmt.Println(time.Now(), CacheDbRef.ID, "start to make FB")
		go SendFinalBlock(&shard.GlobalGroupMems)
	} else {
		CacheDbRef.Mu.Lock()
		CacheDbRef.GenerateTxBlock(0)
		fmt.Println(time.Now(), CacheDbRef.ID, "sends the last TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height", CacheDbRef.TxB.Height, " Bad")
		tmpStr := fmt.Sprintln("Shard", CacheDbRef.ShardNum, ":", CacheDbRef.ID, "sends bad last TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height:", CacheDbRef.TxB.Height)
		sendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
		data3 := new([]byte)
		CacheDbRef.TxB.Encode(data3, 0)
		go SendTxBlock(data3, 0, gVar.GossipRound)
		CacheDbRef.Mu.Unlock()
		//topGetTx <- true
		RollingProcess(false, true, CacheDbRef.TxB)
	}
}
*/

//TxNormalBlock is the loop of TxBlock
func TxNormalBlock(round uint32) {
	if round > 0 {
		<-TBBChan[round-1]
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.GenerateTxBlock(1)
	fmt.Println(time.Now(), CacheDbRef.ID, "sends a TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height:", CacheDbRef.TxB.Height)
	tmpStr := fmt.Sprintln("Shard", CacheDbRef.ShardNum, ":", CacheDbRef.ID, "sends a TxBlock with", CacheDbRef.TxB.TxCnt, "Txs, Height:", CacheDbRef.TxB.Height)
	sendTxMessage(gVar.MyAddress, "LogInfo", []byte(tmpStr))
	/*if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
		fmt.Println(CacheDbRef.ID, "start to make repBlock")
		for i := CacheDbRef.TxB.Height - gVar.NumTxBlockForRep - CacheDbRef.PrevHeight; i < CacheDbRef.TxB.Height-CacheDbRef.PrevHeight; i++ {
			//fmt.Println("Rep prepare: Round", i)
			//fmt.Println(CacheDbRef.RepCache[i])
			for j := uint32(0); j < gVar.ShardSize; j++ {
				shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][j]].Rep += CacheDbRef.RepCache[i][j]
			}
		}
		tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
		tmp := make([][32]byte, gVar.NumTxBlockForRep)
		copy(tmp, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
		CurrentRepRound++
		go LeaderCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: true, Hash: tmp, Rep: tmpRep, Round: CurrentRepRound})
		//startRep <- repInfo{Last: true, Hash: tmp, Rep: tmpRep}
	}*/
	data3 := new([]byte)
	CacheDbRef.TxB.Encode(data3, 0)
	CacheDbRef.TBBCache[round] = CacheDbRef.TxB
	go SendTxBlock(data3, 0, gVar.GossipRound)
	/*if CacheDbRef.TxB.Height == CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
		go TxLastBlock()
	}*/
	CacheDbRef.Mu.Unlock()
	if round < gVar.NumTxListPerEpoch-1 {
		TBBChan[round] <- 1
	}
}

//SendTxList is sending txlist
func SendTxList(data []byte, level uint32) {
	var tmp basic.TxList
	data1 := make([]byte, len(data))
	copy(data1, data)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding the sendTxlist")
	}
	tmp.Sender = CacheDbRef.ID
	for i := level; i > 0; i-- {
		tmp.Level = i
		dest := CacheDbRef.GetGossipID(0, nil)
		data2 := new([]byte)
		tmp.Encode(data2)
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][dest]
		sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxList", *data2)
	}
	/*for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxList", data)
		}
	}*/
}

//SendTxDecSetInShard is sending txDecSet to inshard member
func SendTxDecSetInShard(data []byte, round uint32, level uint32) {
	var tmp basic.TxDecSet
	data1 := make([]byte, len(data))
	copy(data1, data)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Error in decoding the sendTxlist")
	}
	tmp.Sender = CacheDbRef.ID
	for i := level; i > 0; i-- {
		tmp.Level = i
		dest := CacheDbRef.GetGossipID(0, nil)
		data2 := new([]byte)
		tmp.Encode(data2)
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][dest]
		sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", *data2)
	}
	/*for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			//fmt.Println(CacheDbRef.ID, "send TDS to", xx)
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", data[CacheDbRef.ShardNum])
		}
	}*/
}

//SendTxDecSet is sending txDecSet
func SendTxDecSet(data [][]byte, round uint32, level uint32) {
	go SendTxDecSetInShard(data[CacheDbRef.ShardNum], round, level)
	/*rand.Seed(int64(CacheDbRef.ID)*time.Now().Unix() + rand.Int63())
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		xx := rand.Int()%(int(gVar.ShardSize)-1) + 1
		if i != CacheDbRef.ShardNum {
			fmt.Println(CacheDbRef.ID, "(Leader) send TDS to", shard.ShardToGlobal[i][xx], "Shard: ", i, "Its Leader is:", shard.ShardToGlobal[i][0])
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxDecSet", data[i])
		}
	}

	cnt := 1
	mask := make([]bool, gVar.ShardCnt)
	mask[CacheDbRef.ShardNum] = true
	for cnt < int(gVar.ShardCnt) {
		select {
		case tmp := <-TxDecRevChan[round]:
			fmt.Println("Get txdecRev from", tmp.ID)
			mask[tmp.ID] = true
			cnt++
		case <-time.After(timeoutTxDecRev):
			for i := uint32(0); i < gVar.ShardCnt; i++ {
				if !mask[i] {
					xx := rand.Int()%(int(gVar.ShardSize)-1) + 1
					fmt.Println(CacheDbRef.ID, "(Leader) RE send TDS to", shard.ShardToGlobal[i][xx], "Shard: ", i, "Its Leader is:", shard.ShardToGlobal[i][0])
					sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxDecSet", data[i])
				}
			}
		}
	}*/
}

//SendTxBlock is sending txBlock
func SendTxBlock(data *[]byte, kind int, level uint32) {
	var tmp basic.TxBlock
	data1 := make([]byte, len(*data))
	copy(data1, *data)
	err := tmp.Decode(&data1, kind)
	if err != nil {
		fmt.Println("Error in decoding the sendTxlist")
	}
	tmp.Sender = CacheDbRef.ID
	for i := level; i > 0; i-- {
		tmp.Level = i
		dest := CacheDbRef.GetGossipID(0, nil)
		data2 := new([]byte)
		tmp.Encode(data2, kind)
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][dest]
		sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxB", *data2)
	}
	/*for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxB", *data)
		}
	}*/
}

//HandleTotalTx process the tx
func HandleTotalTx(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	var tmp TxBatchInfo
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Decode tx batch error")
		return err
	}
	//TxBatchCache <- tmp
	return nil
}

//HandleTxMM process the tx
func HandleTxMM(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	var tmp TxBatchInfo
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("TxMM decode error")
		return err
	}
	xxx := txDecRev{ID: CacheDbRef.ID, Round: tmp.Round}
	fmt.Println("Get TxBatchMM, Round", tmp.Round, "from", tmp.ID, "Shard", shard.GlobalGroupMems[tmp.ID].Shard)
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxMMRec", xxx.Encode())
	HandleTotalTx(data1)
	return nil
}

//HandleAndSendTx when receives a tx
func HandleAndSendTx(data []byte) error {
	HandleTotalTx(data)
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxM", data)
		}
	}
	return nil
}

//HandleTxLeader when receives a tx
func HandleTxLeader() {
	flag := true
	var TBCache []*basic.TransactionBatch
	//fmt.Println("Leader starts to get txs")
	for flag {
		select {
		case data := <-TxBatchCache:
			if data.Epoch == uint32(CurrentEpoch+1) {
				data1 := make([]byte, len(data.Data))
				copy(data1, data.Data)
				tmp := new(basic.TransactionBatch)
				err := tmp.Decode(&data1)
				//fmt.Println("Get a batch")
				if err == nil {
					//fmt.Println("Batch is good", tmp.TxCnt)
					TBCache = append(TBCache, tmp)
				}
			}
		case <-time.After(timeoutGetTx):
			if len(TBCache) > 0 {
				CacheDbRef.Mu.Lock()
				//fmt.Println(time.Now(), "TxBatch Started", len(TBCache), "in total")
				tmpCnt := 0
				bad := 0
				for j := 0; j < len(TBCache); j++ {
					tmpCnt += int(TBCache[j].TxCnt)
					for i := uint32(0); i < TBCache[j].TxCnt; i++ {
						err := CacheDbRef.MakeTXList(&TBCache[j].TxArray[i])
						if err != nil {
							bad++
							//fmt.Println(CacheDbRef.ID, "has a error(TxBatch)", i, ": ", err)
						}
					}
					if CacheDbRef.Now.TLS[CacheDbRef.ShardNum].TxCnt > gVar.ShardSize*gVar.NumOfTxForTest {
						CacheDbRef.Mu.Unlock()
						time.Sleep(gVar.TxSendInterval * time.Second)
						CacheDbRef.Mu.Lock()
					}
				}
				fmt.Println(time.Now(), "TxBatch Finished Total:", tmpCnt, "Bad: ", bad)
				CacheDbRef.Mu.Unlock()
				TBCache = make([]*basic.TransactionBatch, 0)
			}
			//case <-StopGetTx:
			//	flag = false
		}
	}
	//fmt.Println("Leader stops to get txs")
}

//HandleTxDecLeader when receives a txdec
func HandleTxDecLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecision)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err)
		return err
	}
	//fmt.Println("Into TxDecLeader func")
	CacheDbRef.Mu.Lock()
	//fmt.Println("Ready to preprocess TxDec")
	err = CacheDbRef.PreTxDecision(tmp, tmp.HashID)
	//fmt.Println("Preprocess TxDec done")
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err, "from", tmp.ID)
	}
	//tmp.Print()
	//fmt.Println(time.Now(), CacheDbRef.ID, "(Leader) get TxDec From", tmp.ID, "Hash: ", base58.Encode(tmp.HashID[:]))
	var x uint32
	err = CacheDbRef.UpdateTXCache(tmp, &x)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDec)", err, "from", tmp.ID)
	}
	CacheDbRef.Mu.Unlock()
	//fmt.Println("TxDecRound:", x)
	TLChan[x] <- tmp.ID
	//fmt.Println("TxDecSignal sent")
	return nil
}

//HandleTxDecRev is handle the receive data from other shard miner
func HandleTxDecRev(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(txDecRev)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error(TxDecRev)", err)
		return err
	}
	TxDecRevChan[tmp.Round-CacheDbRef.PrevHeight] <- *tmp
	return nil
}

//HandleTxDecSetLeader when receives a txdecset
func HandleTxDecSetLeader(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}
	flag := true
	fmt.Println(time.Now(), "Leader", CacheDbRef.ID, "get TDS from", tmp.ID, "with", tmp.TxCnt, "Txs Shard", tmp.ShardIndex, "Round", tmp.Round)
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	if s.Stat == 0 {
		flag = false
	}
	CacheDbRef.Mu.Unlock()
	fmt.Println("TDS of Shard", tmp.ShardIndex, "Round", tmp.Round, "Stat:", s.Stat)
	for flag {
		time.Sleep(time.Microsecond * gVar.GeneralSleepTime)
		CacheDbRef.Mu.Lock()
		if s.Stat == 0 {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
	}
	flag = false
	fmt.Println("TDS of Shard", tmp.ShardIndex, "Round", tmp.Round, "New Stat:", s.Stat)
	CacheDbRef.Mu.Lock()
	fmt.Println(time.Now(), "Leader", CacheDbRef.ID, "get TDS done from", tmp.ID, "with", tmp.TxCnt, "Txs")
	CacheDbRef.ProcessTDS(tmp, nil)
	fmt.Println(time.Now(), "TDS from", tmp.ID, "Done")
	CacheDbRef.TDSCnt[tmp.ShardIndex]++
	if CacheDbRef.TDSCnt[tmp.ShardIndex] == gVar.NumTxListPerEpoch {
		CacheDbRef.TDSNotReady--
		if CacheDbRef.TDSNotReady == 0 {
			fmt.Println("All TDS received")
			flag = true
		}
	}
	CacheDbRef.Mu.Unlock()
	if flag {
		StartLastTxBlock <- CurrentEpoch
	}
	return nil
}

/*--------------Client------------*/

//HandleRequestTxB query the TxBlock
func HandleRequestTxB(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(TxBRequestInfo)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	txBs := CacheDbRef.DB.RecentBlock(uint32(tmp.Height))
	data2 := make([]byte, 0)
	basic.Encode(&data2, len(*txBs))
	for i := len(*txBs) - 1; i >= 0; i-- {
		data2 = append(data2, (*txBs)[i].Serial()...)
	}
	sendTxMessage(tmp.Address, "TxBs", data2)
	return nil
}

//Encode is encode
func (a *TxBRequestInfo) Encode() []byte {
	tmp := make([]byte, 0, 12+len(a.Address))
	basic.Encode(&tmp, []byte(a.Address))
	basic.Encode(&tmp, a.Height)
	basic.Encode(&tmp, a.Shard)
	return tmp
}

//Decode is encode
func (a *TxBRequestInfo) Decode(buf *[]byte) error {
	var xxx []byte
	err := basic.Decode(buf, &xxx)
	if err != nil {
		return err
	}
	a.Address = string(xxx)
	err = basic.Decode(buf, &a.Height)
	if err != nil {
		return err
	}
	err = basic.Decode(buf, &a.Shard)
	if err != nil {
		return err
	}
	return nil
}

//Encode is encode
func (a *txDecRev) Encode() []byte {
	tmp := make([]byte, 0, 8)
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.Round)
	return tmp
}

//Decode is encode
func (a *txDecRev) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return err
	}

	err = basic.Decode(buf, &a.Round)
	if err != nil {
		return err
	}
	return nil
}
