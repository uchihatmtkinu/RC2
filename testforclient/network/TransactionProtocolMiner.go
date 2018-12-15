package network

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/uchihatmtkinu/RC/base58"
	"github.com/uchihatmtkinu/RC/basic"
	"github.com/uchihatmtkinu/RC/gVar"
	"github.com/uchihatmtkinu/RC/rccache"
	"github.com/uchihatmtkinu/RC/shard"
)

//MinerLoop is the loop of miner
func MinerLoop() {
	for i := 0; i < gVar.NumTxListPerEpoch; i++ {
		time.Sleep(gVar.TxSendInterval * time.Second)
		go HandleTxLoopMiner(uint32(i))
	}
}

//HandleTxLoopMiner is to handle the tx logic
func HandleTxLoopMiner(round uint32) {
	flag := true
	for flag {
		CacheDbRef.Mu.Lock()
		if CacheDbRef.TLSCache[round] == nil {
			go QueryTL(round)
		} else {
			flag = false
		}
		CacheDbRef.Mu.Unlock()
		time.Sleep(time.Second)
	}
	time.Sleep(gVar.TxSendInterval * time.Second)
	flag1 := true
	flag2 := true
	for flag1 || flag2 {
		CacheDbRef.Mu.Lock()
		if CacheDbRef.TDSCache[round] == nil {
			go QueryTDS(round)
		} else {
			flag1 = false
		}
		if CacheDbRef.TBBCache[round] == nil {
			go QueryTB(round)
		} else {
			flag2 = false
		}
		CacheDbRef.Mu.Unlock()
		time.Sleep(time.Second)
	}
}

//QueryTL queries the txdecset
func QueryTL(round uint32) {
	ran := rand.Intn(int(gVar.ShardSize))
	for ran == int(CacheDbRef.ID) {
		ran = rand.Intn(int(gVar.ShardSize))
	}
	tmpdata := QTL{CacheDbRef.ID, round}
	tmp := tmpdata.Encode()

	sendTxMessage(shard.GlobalGroupMems[ran].Address, "QueryTL", tmp)
}

//HandleQTL is handle
func HandleQTL(data []byte) {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(QTL)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("QTL error ")
	}
	CacheDbRef.Mu.Lock()
	if CacheDbRef.TLSCache[tmp.Round] != nil {
		tmpData := []byte{}
		(*CacheDbRef.TLSCache[tmp.Round]).Level = 0
		(*CacheDbRef.TLSCache[tmp.Round]).Sender = CacheDbRef.ID
		(*CacheDbRef.TLSCache[tmp.Round]).Encode(&tmpData)
		sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxList", tmpData)
	}
	CacheDbRef.Mu.Unlock()
}

//HandleQTDS is handle
func HandleQTDS(data []byte) {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(QTDS)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("QTDS error ")
	}
	CacheDbRef.Mu.Lock()
	if CacheDbRef.TDSCache[tmp.Round] != nil {
		tmpData := []byte{}
		(*CacheDbRef.TDSCache[tmp.Round]).Level = 0
		(*CacheDbRef.TDSCache[tmp.Round]).Sender = CacheDbRef.ID
		(*CacheDbRef.TDSCache[tmp.Round]).Encode(&tmpData)
		sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDecSetM", tmpData)
	}
	CacheDbRef.Mu.Unlock()
}

//HandleQTB is handle
func HandleQTB(data []byte) {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(QTB)
	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("QTB error ")
	}
	CacheDbRef.Mu.Lock()
	if CacheDbRef.TBBCache[tmp.Round] != nil {
		tmpData := []byte{}
		(*CacheDbRef.TBBCache[tmp.Round]).Level = 0
		(*CacheDbRef.TBBCache[tmp.Round]).Sender = CacheDbRef.ID
		(*CacheDbRef.TBBCache[tmp.Round]).Encode(&tmpData, 0)
		sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxB", tmpData)
	}
	CacheDbRef.Mu.Unlock()
}

//QueryTDS queries the txdecset
func QueryTDS(round uint32) {
	ran := rand.Intn(int(gVar.ShardSize))
	for ran == int(CacheDbRef.ID) {
		ran = rand.Intn(int(gVar.ShardSize))
	}
	tmpdata := QTDS{CacheDbRef.ID, round}
	tmp := tmpdata.Encode()
	sendTxMessage(shard.GlobalGroupMems[ran].Address, "QueryTDS", tmp)
}

//QueryTB queries the txdecset
func QueryTB(round uint32) {
	ran := rand.Intn(int(gVar.ShardSize))
	for ran == int(CacheDbRef.ID) {
		ran = rand.Intn(int(gVar.ShardSize))
	}
	tmpdata := QTB{CacheDbRef.ID, round}
	tmp := tmpdata.Encode()
	sendTxMessage(shard.GlobalGroupMems[ran].Address, "QueryTB", tmp)
}

//HandleTx when receives a tx
func HandleTx() {
	flag := true
	sendFlag := false
	var TBCache []*basic.TransactionBatch
	for flag {
		select {
		case data := <-TxBatchCache:
			if data.Epoch == uint32(CurrentEpoch+1) {
				data1 := make([]byte, len(data.Data))
				copy(data1, data.Data)
				tmp := new(basic.TransactionBatch)
				err := tmp.Decode(&data1)
				if err == nil {
					TBCache = append(TBCache, tmp)
				}
				if data.ID == CacheDbRef.Leader && !sendFlag {
					fmt.Println("Start sending packets")
					StartSendingTx <- true
					sendFlag = true
				}
			}
		case <-time.After(timeoutGetTx):
			if len(TBCache) > 0 {
				CacheDbRef.Mu.Lock()
				tmpCnt := 0
				bad := 0
				//fmt.Println(time.Now(), "TxBatch Started", len(TBCache), "in total")
				for j := 0; j < len(TBCache); j++ {
					tmpCnt += int(TBCache[j].TxCnt)
					for i := uint32(0); i < TBCache[j].TxCnt; i++ {
						err := CacheDbRef.GetTx(&TBCache[j].TxArray[i])
						if err != nil {
							bad++
							//fmt.Println(CacheDbRef.ID, "has a error(TxBatch)", i, ": ", err)
						}
					}
				}
				//fmt.Println(time.Now(), "TxBatch Finished Total:", tmpCnt, "Bad: ", bad)
				CacheDbRef.Mu.Unlock()
				TBCache = make([]*basic.TransactionBatch, 0)
			}
			//case <-StopGetTx:
			//flag = false
		}
	}

}

//HandleTxList when receives a txlist
func HandleTxList(data []byte) error {

	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxList)
	err := tmp.Decode(&data1)
	if err != nil {
		return err
	}
	//fmt.Println(CacheDbRef.ID, "get TxList from", tmp.ID)
	//fmt.Println("StropGetTx", CacheDbRef.StopGetTx, "TLRound:", CacheDbRef.TLRound, "tmpRound:", tmp.Round)
	fmt.Println(time.Now(), CacheDbRef.ID, "gets a txlist with", tmp.TxCnt, "Txs", "Current round:", CacheDbRef.TLRound, "its round", tmp.Round, base58.Encode(tmp.HashID[:]))
	if CacheDbRef.TLCheck[tmp.Round-CacheDbRef.PrevHeight] {
		return nil
	}
	CacheDbRef.TLCheck[tmp.Round-CacheDbRef.PrevHeight] = true
	data2 := make([]byte, len(data))
	tmp.Sender = CacheDbRef.ID
	tmp.Encode(&data2)
	go SendTxList(data2, tmp.Level-1)
	/*s := rccache.PreStat{Stat: -2, Valid: nil}
	CacheDbRef.Mu.Lock()
	//fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]))
	CacheDbRef.PreTxList(tmp, &s)
	//fmt.Println(time.Now(), "PreProcess TxList:", base58.Encode(tmp.HashID[:]), "Done")
	CacheDbRef.Mu.Unlock()
	if s.Stat != 0 {
		fmt.Println("TxList:", base58.Encode(tmp.HashID[:]), "Need waiting, remaining:", s.Stat)
	}
	timeoutFlag := true
	cnt := s.Stat*/
	/*for timeoutFlag && s.Stat > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutTL):
			fmt.Println(time.Now(), "TxList:", base58.Encode(tmp.HashID[:]), "time out")
			timeoutFlag = false
		}
	}*/
	/*if s.Stat == 0 {
		timeoutFlag = false
	}
	for timeoutFlag {
		time.Sleep(time.Microsecond * gVar.GeneralSleepTime * 4)
		CacheDbRef.Mu.Lock()
		if s.Stat == 0 {
			timeoutFlag = false
		}
		CacheDbRef.Mu.Unlock()
	}

	fmt.Println("TxList Round", tmp.Round, "Stat: ", s.Stat, "PrevHeight:", CacheDbRef.PrevHeight)*/
	//fmt.Println(time.Now(), "Start Process TxList", base58.Encode(tmp.HashID[:]))
	CacheDbRef.Mu.Lock()
	tmpBatch := new([]basic.TransactionBatch)
	err = CacheDbRef.ProcessTL(tmp, tmpBatch)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
	var sent []byte
	CacheDbRef.TLNow.Encode(&sent)
	CacheDbRef.Mu.Unlock()
	fmt.Println(time.Now(), "Start Sending TxBatch to other shards", base58.Encode(tmp.HashID[:]))
	sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDec", sent)
	//xx := shard.MyMenShard.InShardId
	//thisRound := tmp.Round - CacheDbRef.PrevHeight
	//BatchCache[thisRound] = make([]TxBatchInfo, gVar.ShardCnt)
	/*yy := -1
	for i := uint32(0); i < gVar.ShardCnt; i++ {
		BatchCache[thisRound][i].Data = (*tmpBatch)[i].Encode()
		BatchCache[thisRound][i].ID = CacheDbRef.ID
		BatchCache[thisRound][i].ShardID = CacheDbRef.ShardNum
		BatchCache[thisRound][i].Epoch = uint32(CurrentEpoch + 1)
		BatchCache[thisRound][i].Round = tmp.Round
		if i != CacheDbRef.ShardNum {
			fmt.Println("Send TxBatch, Round", tmp.Round, "to", shard.ShardToGlobal[i][xx], "Shard", i)
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxMM", BatchCache[thisRound][i].Encode())
			if xx == int(i+1) {
				yy = int(i)
				fmt.Println("Send TxBatch, Round", tmp.Round, "to Leader", shard.ShardToGlobal[i][0], "Shard", i)
				sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][0]].Address, "TxMM", BatchCache[thisRound][i].Encode())
			}
		}
	}

	mask := make([]bool, gVar.ShardCnt)
	cnt = int(gVar.ShardCnt)
	if yy == -1 {
		mask[CacheDbRef.ShardNum] = true
		cnt--
	}

	for cnt > 0 {
		select {
		case nowInfo := <-txMCh[thisRound]:
			if shard.GlobalGroupMems[nowInfo.ID].Role == 0 {
				mask[CacheDbRef.ShardNum] = true
			} else {
				mask[shard.GlobalGroupMems[nowInfo.ID].Shard] = true
			}
			cnt--
		case <-time.After(timeoutResentTxmm):
			fmt.Println("Resend TxDec", cnt)
			for i := 0; i < len(mask); i++ {
				if !mask[i] {
					if i == int(CacheDbRef.ShardNum) {
						fmt.Println("Resend TxBatch, Round", tmp.Round, "to Leader", shard.ShardToGlobal[yy][0], "Shard", yy)
						sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[yy][0]].Address, "TxMM", BatchCache[thisRound][i].Encode())
					} else {
						fmt.Println("Resend TxBatch, Round", tmp.Round, "to", shard.ShardToGlobal[i][xx], "Shard", i)
						sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[i][xx]].Address, "TxMM", BatchCache[thisRound][i].Encode())
					}
				}
			}
		}
	}*/
	return nil
}

//HandleTxDecSet when receives a txdecset
func HandleTxDecSet(data []byte, typeInput int) error {
	for !startDone {
		time.Sleep(100 * time.Microsecond)
	}
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxDecSet)
	err := tmp.Decode(&data1)
	fmt.Println("Get the tds from leader:", tmp.ID, "Round:", tmp.Round)
	if CacheDbRef.TDSCheck[tmp.Round-CacheDbRef.PrevHeight] {
		return nil
	}
	CacheDbRef.TDSCheck[tmp.Round-CacheDbRef.PrevHeight] = true
	data2 := make([]byte, len(data))
	tmp.Sender = CacheDbRef.ID
	tmp.Encode(&data2)
	go SendTxDecSetInShard(data2, tmp.Round-CacheDbRef.PrevHeight, tmp.Level-1)
	if tmp.Round < CacheDbRef.PrevHeight {
		return fmt.Errorf("Previous epoch packet")
	}
	if typeInput == 1 {
		var tmp1 txDecRev
		tmp1.ID = CacheDbRef.ShardNum
		tmp1.Round = tmp.Round
		datax := tmp1.Encode()
		go sendTxMessage(shard.GlobalGroupMems[tmp.ID].Address, "TxDecRev", datax)
	}
	if err != nil {
		return err
	}
	s := rccache.PreStat{Stat: -2, Valid: nil}

	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxDecSet(tmp, &s)
	CacheDbRef.Mu.Unlock()
	fmt.Println("TxDecSet Round", tmp.Round, "Stat: ", s.Stat)
	if s.Stat > 0 {
		xx := shard.ShardToGlobal[tmp.ShardIndex][rand.Int()%int(gVar.ShardSize-1)+1]
		yy := txDecRev{ID: CacheDbRef.ID, Round: tmp.Round}
		go sendTxMessage(shard.GlobalGroupMems[xx].Address, "RequestTxMM", yy.Encode())
	}
	timeoutFlag := true
	cnt := s.Stat
	cntTimeout := 0
	for timeoutFlag && s.Stat > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutResentTxmm):
			if cntTimeout == 5 {
				fmt.Println("TDS of", tmp.ID, "Round", tmp.Round, "time out")
				timeoutFlag = false
			} else {
				xx := shard.ShardToGlobal[tmp.ShardIndex][rand.Int()%int(gVar.ShardSize-1)+1]
				yy := txDecRev{ID: CacheDbRef.ID, Round: tmp.Round}
				fmt.Println("Request TDS of", tmp.ID, "Round", tmp.Round, "from", xx)
				go sendTxMessage(shard.GlobalGroupMems[xx].Address, "RequestTxMM", yy.Encode())
				cntTimeout++
			}
		}
	}
	fmt.Println("TxDecSet Round", tmp.Round, "New Stat: ", s.Stat)
	if tmp.Round < gVar.NumTxListPerEpoch+CacheDbRef.PrevHeight && tmp.ShardIndex == CacheDbRef.ShardNum && tmp.ID == CacheDbRef.Leader {
		TDSChan[tmp.Round-CacheDbRef.PrevHeight] <- CurrentEpoch
	}
	tmpflag := false
	CacheDbRef.Mu.Lock()
	CacheDbRef.TDSCnt[tmp.ShardIndex]++
	fmt.Println(time.Now(), "Miner", CacheDbRef.ID, "get TDS from", tmp.ID, "with", tmp.TxCnt, "Txs Shard", tmp.ShardIndex, "Round", tmp.Round)
	err = CacheDbRef.GetTDS(tmp, &CacheDbRef.RepCache[tmp.Round-CacheDbRef.PrevHeight])
	fmt.Println(time.Now(), "Miner", CacheDbRef.ID, "get TDS done from", tmp.ID, "with", tmp.TxCnt, "Txs Shard", tmp.ShardIndex, "Round", tmp.Round)
	if err != nil {
		fmt.Println(CacheDbRef.ID, "has a error", err)
	}
	if CacheDbRef.TDSCnt[tmp.ShardIndex] == gVar.NumTxListPerEpoch {
		CacheDbRef.TDSNotReady--
		fmt.Println("Decrease the TDSCnt to", CacheDbRef.TDSNotReady)
		if CacheDbRef.TDSNotReady == 0 {
			tmpflag = true
		}
	}
	CacheDbRef.Mu.Unlock()
	if tmpflag {
		//StopGetTx <- true
		StartLastTxBlock <- CurrentEpoch
	}
	return nil
}

//HandleAndSentTxDecSet when receives a txdecset
func HandleAndSentTxDecSet(data []byte) error {
	err := HandleTxDecSet(data, 1)
	if err != nil {
		return err
	}
	fmt.Println(CacheDbRef.ID, "Get TDS and send")
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxDecSetM", data)
		}
	}
	return nil
}

//HandleTxBlock when receives a txblock
func HandleTxBlock(data []byte) error {

	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(basic.TxBlock)
	err := tmp.Decode(&data1, 0)
	if err != nil {
		fmt.Println("Error in decoding txblock", err)
		return err
	}
	if CacheDbRef.TBCheck[tmp.Height-CacheDbRef.PrevHeight] {
		return nil
	}
	CacheDbRef.TBCheck[tmp.Height-CacheDbRef.PrevHeight] = true
	data2 := make([]byte, len(data))
	tmp.Sender = CacheDbRef.ID
	tmp.Encode(&data2, 0)
	go SendTxDecSetInShard(data2, tmp.Height-CacheDbRef.PrevHeight, tmp.Level-1)
	s := rccache.PreStat{Stat: -2, Valid: nil}
	if tmp.Height < CacheDbRef.PrevHeight {
		return fmt.Errorf("Previous epoch txblock")
	}
	CacheDbRef.Mu.Lock()
	CacheDbRef.PreTxBlock(tmp, &s)
	CacheDbRef.Mu.Unlock()

	timeoutFlag := true
	cnt := s.Stat
	for timeoutFlag && s.Stat > 0 {
		select {
		case <-s.Channel:
			cnt--
		case <-time.After(timeoutTL):
			timeoutFlag = false
		}
	}
	if cnt == 0 {
		fmt.Println("Get txBlock from", tmp.ID, "Hash:", base58.Encode(tmp.HashID[:]), "preprocess done")
	} else {
		fmt.Println("Get txBlock from", tmp.ID, "Hash:", base58.Encode(tmp.HashID[:]), "preprocess timeout")
	}

	if tmp.Height <= CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
		waitFlag := true
		for waitFlag {
			tmpInt := <-TDSChan[tmp.Height-CacheDbRef.PrevHeight-1]
			if tmpInt == CurrentEpoch {
				waitFlag = false
			}
		}
	}
	if tmp.Height <= CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch+1 && tmp.Height >= CacheDbRef.PrevHeight+2 {
		waitFlag := true
		for waitFlag {
			tmpInt := <-TBChan[tmp.Height-CacheDbRef.PrevHeight-2]
			if tmpInt == CurrentEpoch {
				waitFlag = false
			}
		}
	}
	if tmp.Height == CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch+1 {
		EpochFlag := true
		for EpochFlag {
			tmpEpoch := <-StartLastTxBlock
			if tmpEpoch == CurrentEpoch {
				EpochFlag = false
			}
		}
	}
	flag := true
	fmt.Println("TxB Kind", tmp.Kind)
	if tmp.Kind != 3 {
		for flag {
			CacheDbRef.Mu.Lock()
			err = CacheDbRef.GetTxBlock(tmp)
			if err != nil {
				//fmt.Println("txBlock", base58.Encode(tmp.HashID[:]), " error", err)
			} else {
				flag = false
			}
			CacheDbRef.Mu.Unlock()
			time.Sleep(time.Microsecond * gVar.GeneralSleepTime)
		}

		CacheDbRef.Mu.Lock()
		fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBlock with", tmp.TxCnt, "Txs from", tmp.ID, "Hash", base58.Encode(tmp.HashID[:]), "Height:", tmp.Height)

		if tmp.Height == CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch+1 {
			fmt.Println(time.Now(), CacheDbRef.ID, "waits for FB")
			/*(for i := tmp.Height - uint32(len(*CacheDbRef.TBCache)) - CacheDbRef.PrevHeight; i < tmp.Height-1-CacheDbRef.PrevHeight; i++ {
				//fmt.Println("Rep prepare: Round", i)
				//fmt.Println(CacheDbRef.RepCache[i])
				for j := uint32(0); j < gVar.ShardSize; j++ {
					shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][j]].Rep += CacheDbRef.RepCache[i][j]
				}
			}
			tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
			tmpHash := make([][32]byte, len(*CacheDbRef.TBCache))
			//copy(tmpHash, *CacheDbRef.TBCache)
			//*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[len(*CacheDbRef.TBCache):]
			//CurrentRepRound++
			//fmt.Println(time.Now(), CacheDbRef.ID, "start to make last repBlock, Round:", CurrentRepRound)

			//go MemberCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: false, Hash: tmpHash, Rep: tmpRep, Round: CurrentRepRound})
			*/
			//		go WaitForFinalBlock(&shard.GlobalGroupMems)
		} else {
			/*if len(*CacheDbRef.TBCache) >= gVar.NumTxBlockForRep {
				fmt.Println(CacheDbRef.ID, "start to make repBlock")
				tmpHash := make([][32]byte, gVar.NumTxBlockForRep)
				copy(tmpHash, (*CacheDbRef.TBCache)[0:gVar.NumTxBlockForRep])
				for i := tmp.Height - gVar.NumTxBlockForRep - CacheDbRef.PrevHeight; i < tmp.Height-CacheDbRef.PrevHeight; i++ {
					//fmt.Println("Rep prepare: Round", i)
					//fmt.Println(CacheDbRef.RepCache[i])
					for j := uint32(0); j < gVar.ShardSize; j++ {
						shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][j]].Rep += CacheDbRef.RepCache[i][j]
					}
				}
				tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
				*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[gVar.NumTxBlockForRep:]
				CurrentRepRound++
				go MemberCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: true, Hash: tmpHash, Rep: tmpRep, Round: CurrentRepRound})
			}*/
		}
		CacheDbRef.Mu.Unlock()
		if tmp.Height <= CacheDbRef.PrevHeight+gVar.NumTxListPerEpoch {
			TBChan[tmp.Height-CacheDbRef.PrevHeight-1] <- CurrentEpoch
		}
	} else {
		fmt.Println(time.Now(), CacheDbRef.ID, "gets a bad txBlock with", tmp.TxCnt, "Txs from", tmp.ID, "Hash", base58.Encode(tmp.HashID[:]), "Height:", tmp.Height)
		//RollingProcess(true, false, tmp)
	}
	return nil
}

//Encode is encode
func (a *QTL) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.Round)
	return tmp
}

//Decode is decode
func (a *QTL) Decode(data *[]byte) error {
	data1 := make([]byte, len(*data))
	copy(data1, *data)
	err := basic.Decode(&data1, &a.ID)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Round)
	if err != nil {
		return err
	}
	return nil
}

//Encode is encode
func (a *QTB) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.Round)
	return tmp
}

//Decode is decode
func (a *QTB) Decode(data *[]byte) error {
	data1 := make([]byte, len(*data))
	copy(data1, *data)
	err := basic.Decode(&data1, &a.ID)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Round)
	if err != nil {
		return err
	}
	return nil
}

//Encode is encode
func (a *QTDS) Encode() []byte {
	var tmp []byte
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.Round)
	return tmp
}

//Decode is decode
func (a *QTDS) Decode(data *[]byte) error {
	data1 := make([]byte, len(*data))
	copy(data1, *data)
	err := basic.Decode(&data1, &a.ID)
	if err != nil {
		return err
	}
	err = basic.Decode(&data1, &a.Round)
	if err != nil {
		return err
	}
	return nil
}
