package basic

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math/big"
)

//HashCut returns the part of the hash
func HashCut(x [32]byte) [SHash]byte {
	var tmp [SHash]byte
	copy(tmp[:], x[:SHash])
	return tmp
}

//Encode is the general function of decoding
func Encode(tmp *[]byte, out interface{}) error {
	switch out := out.(type) {
	case *[]byte:
		EncodeByte(tmp, out)
	case *[32]byte:
		EncodeByteL(tmp, out[:], 32)
	case *[SHash]byte:
		EncodeByteL(tmp, out[:], SHash)
	case *RCSign:
		out.SignToData(tmp)
	case *big.Int:
		EncodeBig(tmp, out)
	default:
		EncodeInt(tmp, out)
	}
	return nil
}

//Decode is the general function of decoding
func Decode(tmp *[]byte, out interface{}) error {
	var err error
	switch out := out.(type) {
	case *[]byte:
		err = DecodeByte(tmp, out)
	case *[32]byte:
		xxx := make([]byte, 0, 32)
		err = DecodeByteL(tmp, &xxx, 32)
		if err == nil {
			copy(out[:], xxx[:32])
		}
	case *[SHash]byte:
		xxx := make([]byte, 0, SHash)
		err = DecodeByteL(tmp, &xxx, SHash)
		if err == nil {
			copy(out[:], xxx[:SHash])
		}
	case *RCSign:
		err = out.DataToSign(tmp)
	case *big.Int:
		err = DecodeBig(tmp, out)
	default:
		err = DecodeInt(tmp, out)
	}
	return err
}

//DoubleHash256 returns the double hash result of hash256
func DoubleHash256(a *[]byte, b *[32]byte) {
	*b = sha256.Sum256(*a)
	//*b = sha256.Sum256(tmp[:])
}

//EncodeByte add length bit at the head of the byte array
func EncodeByte(current *[]byte, d *[]byte) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(len(*d)))
	if err != nil {
		return fmt.Errorf("Basic.EncodeByte write failed, %s", err)
	}
	tmp := append(buf.Bytes(), *d...)
	*current = append(*current, tmp...)
	return nil
}

//DecodeByte Decode bytes with a uint32 preamble
func DecodeByte(d *[]byte, out *[]byte) error {
	tmpBit := binary.LittleEndian.Uint32((*d)[:4])
	if 4+tmpBit > uint32(len(*d)) {
		return fmt.Errorf("Basic.DecodeByte length not enough")
	}
	tmpArray := (*d)[4 : 4+tmpBit]
	*d = (*d)[4+tmpBit:]
	*out = tmpArray
	return nil
}

//EncodeByteL add length bit at the head of the byte array with specific length
func EncodeByteL(current *[]byte, d []byte, l int) error {
	if len(d) != l {
		return fmt.Errorf("Basic.EncodeByteL length is not right")
	}
	*current = append(*current, d...)
	return nil
}

//DecodeByteL Decode bytes with a uint32 preamble
func DecodeByteL(d *[]byte, out *[]byte, l int) error {
	if len(*d) < l {
		return fmt.Errorf("Basic.DecodeByteL length not enough")
	}
	*out = (*d)[:l]
	*d = (*d)[l:]
	return nil
}

//byteSlice returns a slice of a integer
func byteSlice(x uint32) []byte {
	var tmp []byte
	EncodeInt(&tmp, x)
	return tmp
}

//ByteSlice returns a slice of a integer
func ByteSlice(x uint32) []byte {
	var tmp []byte
	EncodeInt(&tmp, x)
	return tmp
}

//EncodeInt Encode the data
func EncodeInt(current *[]byte, d interface{}) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, d)
	if err != nil {
		return fmt.Errorf("Basic.EncodeInt write failed, %s", err)
	}
	*current = append(*current, buf.Bytes()...)
	return nil
}

//DecodeInt Decode the data to uint32
func DecodeInt(d *[]byte, out interface{}) error {
	var bitlen uint32

	switch out := out.(type) {
	case *uint32:
		bitlen = 4
		if bitlen > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		tmpBit := binary.LittleEndian.Uint32((*d)[:bitlen])
		*out = tmpBit
	case *int64:
		bitlen = 8
		if bitlen > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		tmpBit := binary.LittleEndian.Uint64((*d)[:bitlen])
		*out = int64(tmpBit)
	case *uint64:
		bitlen = 8
		if bitlen > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		tmpBit := binary.LittleEndian.Uint64((*d)[:bitlen])
		*out = tmpBit
	case *bool:
		bitlen = 1
		if bitlen > uint32(len(*d)) {
			return fmt.Errorf("Basic.DecodeInt length not enough")
		}
		tmpBit := (*d)[0]
		if tmpBit == 0 {
			*out = false
		} else {
			*out = true
		}
	default:
		bitlen = 0
		return fmt.Errorf("Basic.DecodeInt not known type")
	}

	*d = (*d)[bitlen:]

	return nil
}

//EncodeBig encodes a big number into byte[]
func EncodeBig(current *[]byte, d *big.Int) error {
	tmp := uint32(len(d.Bytes()))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, tmp)
	//fmt.Println(tmp)
	if err != nil {
		return fmt.Errorf("Basic.EncodeBig write failed, %s", err)
	}
	*current = append(*current, buf.Bytes()...)
	*current = append(*current, d.Bytes()...)
	return nil
}

//DecodeBig decodes a byte[] into a big number
func DecodeBig(current *[]byte, out *big.Int) error {
	var bitlen uint32
	//fmt.Println(len(*current))
	err := DecodeInt(current, &bitlen)
	//out = new(big.Int)
	if err != nil {
		return err
	}
	//fmt.Println(bitlen)
	if uint32(len(*current)) < bitlen {
		return fmt.Errorf("Basic.DecodeBig not enough length")
	}
	out.SetBytes((*current)[:bitlen][:])
	//fmt.Println(*out)
	*current = (*current)[bitlen:]
	//fmt.Println(len(*current))
	return nil
}

//EncodeDoubleBig encodes a signature into data
func EncodeDoubleBig(current *[]byte, a *big.Int, b *big.Int) error {
	tmp := *current
	err := EncodeBig(&tmp, a)
	if err != nil {
		return err
	}
	err = EncodeBig(&tmp, b)
	if err != nil {
		return err
	}
	*current = tmp
	return nil
}

//DecodeDoubleBig encodes a signature into data
func DecodeDoubleBig(current *[]byte, a *big.Int, b *big.Int) error {
	err := DecodeBig(current, a)
	if err != nil {
		return err
	}
	err = DecodeBig(current, b)
	if err != nil {
		return err
	}
	return nil
}

//Serialize is the general encoding function
func Serialize(a interface{}, b *[]byte) error {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(a)
	if err == nil {
		*b = result.Bytes()
	}
	return err
}

//Deserialize is the general encoding function
func Deserialize(buf *[]byte, a interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(*buf))
	err := decoder.Decode(a)
	return err
}
