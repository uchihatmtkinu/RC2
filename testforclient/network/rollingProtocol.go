package network

/*
//RollingProcess is handle the rolling
func RollingProcess(send bool, FirstLeader bool, TBData *basic.TxBlock) {
	NowSent := send
	Flag := true

	LeaderIndex := 0

	for Flag {
		fmt.Println(time.Now(), "Start Rolling", "Round", LeaderIndex)
		cnt := 0
		if NowSent {
			tmp := rollingInfo{ID: CacheDbRef.ID, Epoch: uint32(CurrentEpoch + 1), Leader: CacheDbRef.Leader}
			rand.Seed(time.Now().Unix() * int64(CacheDbRef.ID))
			tmpx := rand.Perm(int(gVar.ShardSize))
			for i := uint32(0); i < gVar.ShardSize; i++ {
				if shard.ShardToGlobal[CacheDbRef.ShardNum][tmpx[i]] != int(CacheDbRef.ID) {
					//fmt.Println("Send rolling Message to", shard.ShardToGlobal[CacheDbRef.ShardNum][i], tmp.Epoch, tmp.Leader)
					SendRollingMessage(shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][tmpx[i]]].Address, "RollRequest", tmp.Encode())
				}
			}
			cnt = 1
		}

		mask := make([]bool, gVar.ShardSize)
		for i := uint32(0); i < gVar.ShardSize; i++ {
			mask[i] = false
		}
		mask[shard.GlobalGroupMems[MyGlobalID].InShardId] = true
		for cnt < int(gVar.ShardSize)-1-LeaderIndex {
			select {
			case tmpRollMeg := <-rollingChannel:
				xx := shard.GlobalGroupMems[tmpRollMeg.ID].InShardId
				if !mask[xx] && tmpRollMeg.Leader == CacheDbRef.Leader {
					mask[xx] = true
					cnt++
					//fmt.Println(time.Now(), "Get rolling from", tmpRollMeg.ID, xx, "Total", cnt)
				}
			}
		}
		fmt.Println(time.Now(), "Round", LeaderIndex, "New Leader:", shard.ShardToGlobal[CacheDbRef.ShardNum][LeaderIndex+1])

		shard.GlobalGroupMems[CacheDbRef.Leader].ClearTotalRep()
		shard.GlobalGroupMems[CacheDbRef.Leader].ClearRep()
		for i := 0; i < len(CacheDbRef.RepCache); i++ {
			CacheDbRef.RepCache[i][LeaderIndex] = 0
		}
		LeaderIndex++
		CacheDbRef.Leader = uint32(shard.ShardToGlobal[CacheDbRef.ShardNum][LeaderIndex])
		if CacheDbRef.Leader == CacheDbRef.ID {
			maskVD := make([]bool, gVar.ShardSize)
			for i := uint32(0); i < gVar.ShardSize; i++ {
				maskVD[i] = false
			}
			maskVD[shard.GlobalGroupMems[MyGlobalID].InShardId] = true
			tmpCnt := 1
			for tmpCnt < int(gVar.ShardSize) {
				select {
				case tmpVD := <-VTDChannel:
					xx := shard.GlobalGroupMems[tmpVD.ID].InShardId
					if tmpVD.Leader == CacheDbRef.Leader && !maskVD[xx] {
						maskVD[xx] = true
						tmpCnt++
					}
				}
			}
			fmt.Println("Get enough td")
			if CacheDbRef.Badness {
				tmpData := new([]byte)
				TBData.Encode(tmpData, 0)
				go SendVirtualTDS(*tmpData)
				go SendTxBlockAfterRolling(tmpData)
				NowSent = false
			} else {
				TBData.Kind = 0
				tmpData := new([]byte)
				TBData.Encode(tmpData, 0)
				go SendVirtualTDS(*tmpData)
				go SendTxBlockAfterRolling(tmpData)
				Flag = false
			}
		} else {
			tmpVD := rollingInfo{ID: CacheDbRef.ID, Epoch: uint32(CurrentEpoch + 1), Leader: CacheDbRef.Leader}
			sendTxMessage(shard.GlobalGroupMems[CacheDbRef.Leader].Address, "VTD", tmpVD.Encode())
			tmpTxB := <-rollingTxB
			tmp := new(basic.TxBlock)
			err := tmp.Decode(&tmpTxB, 0)
			if err != nil {
				fmt.Print("Rolling TxB decoding error")
			}
			if tmp.Kind == 0 {
				Flag = false
			}
		}
	}
	fmt.Println("Rolling done, new leader", CacheDbRef.Leader)
	LeaderAddr = shard.GlobalGroupMems[shard.ShardToGlobal[shard.MyMenShard.Shard][LeaderIndex]].Address
	if FirstLeader {
		CacheDbRef.Mu.Lock()
		CacheDbRef.GenerateTxBlock(2)
		fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBlock with", TBData.TxCnt, "Txs from", CacheDbRef.Leader, "Height:", TBData.Height)
		for i := CacheDbRef.TxB.Height - uint32(len(*CacheDbRef.TBCache)) - CacheDbRef.PrevHeight; i < CacheDbRef.TxB.Height-1-CacheDbRef.PrevHeight; i++ {
			fmt.Println("Rep prepare: Round", i)
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
		//StartLastTxBlock <- CurrentEpoch
		go MemberCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: false, Hash: tmp, Rep: tmpRep, Round: CurrentRepRound})
		go WaitForFinalBlock(&shard.GlobalGroupMems)
	} else {
		TBData.Kind = 0
		CacheDbRef.Mu.Lock()
		err := CacheDbRef.GetTxBlock(TBData)
		if err != nil {
			fmt.Println(time.Now(), "Receiving txblock error in rolling:", err)
		}
		if CacheDbRef.Leader == CacheDbRef.ID {
			fmt.Println(time.Now(), CacheDbRef.ID, "sends a txBlock with", TBData.TxCnt, "Txs, Height:", TBData.Height)
		} else {
			fmt.Println(time.Now(), CacheDbRef.ID, "gets a txBlock with", TBData.TxCnt, "Txs from", CacheDbRef.Leader, "Height:", TBData.Height)
		}
		for i := CacheDbRef.TxB.Height - uint32(len(*CacheDbRef.TBCache)) - CacheDbRef.PrevHeight; i < CacheDbRef.TxB.Height-1-CacheDbRef.PrevHeight; i++ {
			fmt.Println("Rep prepare: Round", i)
			for j := uint32(0); j < gVar.ShardSize; j++ {
				shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][j]].Rep += CacheDbRef.RepCache[i][j]
			}
		}
		tmpRep := shard.ReturnRepData(CacheDbRef.ShardNum)
		tmpHash := make([][32]byte, len(*CacheDbRef.TBCache))
		copy(tmpHash, *CacheDbRef.TBCache)
		*CacheDbRef.TBCache = (*CacheDbRef.TBCache)[len(*CacheDbRef.TBCache):]
		CacheDbRef.Mu.Unlock()
		CurrentRepRound++
		fmt.Println(time.Now(), CacheDbRef.ID, "start to make last repBlock, Round:", CurrentRepRound)
		if CacheDbRef.Leader == CacheDbRef.ID {
			go LeaderCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: false, Hash: tmpHash, Rep: tmpRep, Round: CurrentRepRound})
			go SendFinalBlock(&shard.GlobalGroupMems)
		} else {
			go MemberCoSiRepProcess(&shard.GlobalGroupMems, repInfo{Last: false, Hash: tmpHash, Rep: tmpRep, Round: CurrentRepRound})
			go WaitForFinalBlock(&shard.GlobalGroupMems)
		}
	}
}

//HandleVirtualTD is handling the VTD
func HandleVirtualTD(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(rollingInfo)

	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("Virtual TD decode error", err)
		return err
	}
	fmt.Println("Get Virtual TD from", tmp.ID, tmp.Epoch, tmp.Leader)
	if tmp.Epoch == uint32(CurrentEpoch+1) {
		VTDChannel <- *tmp
	}
	return nil
}

//SendVirtualTDS is to send a virtual tds
func SendVirtualTDS(data []byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		if shard.ShardToGlobal[CacheDbRef.ShardNum][i] != MyGlobalID {
			sendTxMessage(shard.GlobalGroupMems[shard.ShardToGlobal[CacheDbRef.ShardNum][i]].Address, "VTDS", data)
		}
	}
}

//HandleVirtualTDS does nothing
func HandleVirtualTDS(data []byte) {}

//SendTxBlockAfterRolling is sending txBlock after rolling
func SendTxBlockAfterRolling(data *[]byte) {
	for i := uint32(0); i < gVar.ShardSize; i++ {
		xx := shard.ShardToGlobal[CacheDbRef.ShardNum][i]
		if xx != int(CacheDbRef.ID) {
			sendTxMessage(shard.GlobalGroupMems[xx].Address, "TxBR", *data)
		}
	}
}

//HandleTxBlockAfterRolling is handle the txblock after rolling
func HandleTxBlockAfterRolling(data []byte) {
	rollingTxB <- data
}

//SendRollingMessage is sending the rolling request message
func SendRollingMessage(addr string, command string, message []byte) {
	request := append(commandToBytes(command), message...)
	sendData(addr, request)
}

//HandleRollingMessage is handle the received rolling message
func HandleRollingMessage(data []byte) error {
	data1 := make([]byte, len(data))
	copy(data1, data)
	tmp := new(rollingInfo)

	err := tmp.Decode(&data1)
	if err != nil {
		fmt.Println("RollingMessage decode error", err)
		return err
	}
	fmt.Println("Get rolling Message from", tmp.ID, tmp.Epoch, tmp.Leader)
	if tmp.Epoch == uint32(CurrentEpoch+1) {
		rollingChannel <- *tmp
	}
	return nil
}

//Encode is encode
func (a *rollingInfo) Encode() []byte {
	tmp := make([]byte, 0, 8)
	basic.Encode(&tmp, a.ID)
	basic.Encode(&tmp, a.Epoch)
	basic.Encode(&tmp, a.Leader)
	return tmp
}

//Decode is encode
func (a *rollingInfo) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return err
	}
	err = basic.Decode(buf, &a.Epoch)
	if err != nil {
		return err
	}
	err = basic.Decode(buf, &a.Leader)
	if err != nil {
		return err
	}
	return nil
}
*/
