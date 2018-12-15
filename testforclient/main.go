package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/uchihatmtkinu/RC/basic"
)

type PrevOut struct {
	Addr  string `json:"addr"`
	Value int    `json:"value"`
}

type JsInput struct {
	Prev PrevOut `json:"prev_out"`
}

type JsOut struct {
	Addr  string `json:"addr"`
	Value int    `json:"value"`
}

type JsTx struct {
	Hash   string    `json:"hash"`
	VinSz  int       `json:"vin_sz"`
	VoutSz int       `json:"vout_sz"`
	Input  []JsInput `json:"inputs"`
	Out    []JsOut   `json:"out"`
}

type JsBlock struct {
	Hash      string `json:"hash"`
	Prevblock string `json:"prev_block"`
	Fee       int    `json:"fee"`
	Ntx       int    `json:"n_tx"`
	Height    int    `json:"height"`
	Tx        []JsTx `json:"tx"`
}

type JsData struct {
	Blocks []JsBlock `json:"blocks"`
}

var txCnt int
var flag bool
var transactionCache *[]basic.Transaction

func JsonMake(stat int, start int, last int, name int) int {
	//url := "https://blockchain.info/rawblock/00000000d1145790a8694403d4063f323d499e655c83426834d4ce2f8dd4a2ee"
	var file *os.File
	if stat == 0 {
		file, _ = os.Create("TxData" + strconv.Itoa(name) + ".txt")
	} else {
		file, _ = os.OpenFile("TxData"+strconv.Itoa(name)+".txt", os.O_WRONLY|os.O_APPEND, 0666)
	}
	defer file.Close()
	for i := start; i < last; i++ {
		url := "https://blockchain.info/block-height/" + strconv.Itoa(i) + "?format=json"
		spaceClient := http.Client{
			Timeout: time.Second * 2,
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
		}
		req.Header.Set("User-Agent", "spacecount-tutorial")
		res, getErr := spaceClient.Do(req)
		if getErr != nil {
			//log.Fatal(getErr)
			return i
		}
		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			//log.Fatal(readErr)
			return i
		}
		tmp := JsData{}
		jsonErr := json.Unmarshal(body, &tmp)
		if jsonErr != nil {
			log.Fatal(jsonErr)
			//return i
		}
		tmp2, parErr := json.Marshal(tmp)
		if parErr != nil {
			log.Fatal(parErr)
		}

		//fmt.Println(body)
		file.Write(basic.ByteSlice(uint32(len(tmp2))))
		file.Write(tmp2)

		fmt.Println(i)
	}

	return last
	/*file, _ = os.Open("TxData.txt")
	for i := 100000; i < 100200; i++ {
		tmp1 := make([]byte, 4)
		file.Read(tmp1)
		var lenX uint32
		basic.Decode(&tmp1, &lenX)
		tmp2 := make([]byte, lenX)
		file.Read(tmp2)
		tmp := JsData{}
		jsonErr := json.Unmarshal(tmp2, &tmp)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		fmt.Println(tmp)
	}
	file.Close()*/

}

func getTxDataFromFile() {

}

func main() {
	flag = true
	txCnt = 0
	file, _ := os.Open("TxData4.txt")
	for i := 0; i < 50000; i++ {
		tmp1 := make([]byte, 4)
		file.Read(tmp1)
		var lenX uint32
		basic.Decode(&tmp1, &lenX)
		tmp2 := make([]byte, lenX)
		file.Read(tmp2)
		tmp := JsData{}
		jsonErr := json.Unmarshal(tmp2, &tmp)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		fmt.Println(tmp.Blocks[0].Height)
		txCnt += tmp.Blocks[0].Ntx
	}
	fmt.Println(txCnt)
	file.Close()
}
