package newrep

import (
	"fmt"

	"github.com/uchihatmtkinu/RC/basic"
)

//GossipFirMsgMake is to init a GossipFirMsg
func GossipFirMsgMake(data RepMsg) GossipFirMsg {
	var tmp GossipFirMsg
	tmp.ID = data.ID
	tmp.Cnt = 1
	tmp.Data = make([]RepMsg, 1, 1)
	tmp.Data[0] = data
	return tmp
}

//Add adds a new RepMsg
func (a *GossipFirMsg) Add(data RepMsg) {
	a.Cnt++
	a.Data = append(a.Data, data)
}

//GossipSecMsgMake is to init a GossipSecMsg
func GossipSecMsgMake(data RepSecMsg) GossipSecMsg {
	var tmp GossipSecMsg
	tmp.ID = data.ID
	tmp.Cnt = 1
	tmp.Data = make([]RepSecMsg, 1, 1)
	tmp.Data[0] = data
	return tmp
}

//Add adds a new RepSecMsg
func (a *GossipSecMsg) Add(data RepSecMsg) {
	a.Cnt++
	a.Data = append(a.Data, data)
}

//Encode encode the GossipFirMsg into []byte
func (a *GossipFirMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, a.ID)
	basic.Encode(tmp, a.Cnt)
	for i := uint32(0); i < a.Cnt; i++ {
		a.Data[i].Encode(tmp)
	}
}

//Decode decode the []byte into GossipFirMsg
func (a *GossipFirMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("GossipFirMsg ID decode failed: %s", err)
	}
	err = basic.Decode(buf, &a.Cnt)
	if err != nil {
		return fmt.Errorf("GossipFirMsg Cnt decode failed: %s", err)
	}
	for i := uint32(0); i < a.Cnt; i++ {
		err = a.Data[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("GossipFirMsg Data decode failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("GossipFirMsg decode failed: With extra bits")
	}
	return nil
}

//Encode encode the GossipSecMsg into []byte
func (a *GossipSecMsg) Encode(tmp *[]byte) {
	basic.Encode(tmp, a.ID)
	basic.Encode(tmp, a.Cnt)
	for i := uint32(0); i < a.Cnt; i++ {
		a.Data[i].Encode(tmp)
	}
}

//Decode decode the []byte into GossipSecMsg
func (a *GossipSecMsg) Decode(buf *[]byte) error {
	err := basic.Decode(buf, &a.ID)
	if err != nil {
		return fmt.Errorf("GossipSecMsg ID Read failed: %s", err)
	}
	err = basic.Decode(buf, &a.Cnt)
	if err != nil {
		return fmt.Errorf("GossipSecMsg Cnt Read failed: %s", err)
	}
	for i := uint32(0); i < a.Cnt; i++ {
		err = a.Data[i].Decode(buf)
		if err != nil {
			return fmt.Errorf("GossipSecMsg Data Read failed: %s", err)
		}
	}
	if len(*buf) != 0 {
		return fmt.Errorf("GossipSecMsg decode failed: With extra bits")
	}
	return nil
}
