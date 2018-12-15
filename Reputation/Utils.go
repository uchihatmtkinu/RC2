package Reputation

import (
	"bytes"
	"encoding/binary"
	"log"
	"github.com/uchihatmtkinu/RC/Reputation/cosi"
)



// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// BoolToHex converts a bool to byte array
func BoolToHex(f bool) []byte {
	var a []byte
	if f {
		a = append(a,1)
	} else {
		a = append(a,0)
	}

	return a
}

// maskBit returns a boolean value indicating whether the indicated signer is Enabled or Disabled.
func maskBit(signer int, mask *[]byte) (value cosi.MaskBit) {
	byt := signer >> 3
	bit := byte(1) << uint(signer&7)
	return ((*mask)[byt] & bit) != 0
}

/*
func UIntToHex(num uint64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}*/