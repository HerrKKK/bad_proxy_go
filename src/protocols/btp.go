package protocols

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type BTPRequest struct {
	Address string
	Payload []byte
}

func (request *BTPRequest) Parse(buffer []byte) (req BTPRequest, err error) {
	secret := "test secret"
	confusionLen := int(buffer[32]) // uint8
	if confusionLen < 0 || confusionLen > 64 {
		return req, errors.New("unexpected confusion length " + strconv.Itoa(confusionLen))
	}
	pos := 32 + 1 + int(confusionLen)
	timestamp := binary.BigEndian.Uint32(buffer[pos : pos+4])
	timediff := int(time.Now().Unix()) - int(timestamp)
	if timediff < -210 || 210 < timediff {
		return req, errors.New("timeout or replay, time diff is " + strconv.Itoa(timediff))
	}
	pos += 5 // 4 for timestamp 1 for directives
	hostLen := int(buffer[pos])
	pos++
	host := string(buffer[pos : pos+hostLen])
	pos += hostLen
	port := strconv.Itoa(int(binary.BigEndian.Uint16(buffer[pos : pos+2])))
	pos += 2

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(buffer[32:])
	receivedDigest := buffer[:32]
	_ = receivedDigest // add to lru
	digest := h.Sum(nil)
	if string(digest) != string(buffer[:32]) {
		return req, errors.New("digest mismatch, unauthorized connect")
	}
	request.Address = host + ":" + port
	request.Payload = buffer[pos:]
	return *request, nil
}

func EncodeBtpRequest(address string, payload []byte) (res []byte, err error) {
	secret := "test secret"
	bytesBuffer := bytes.NewBuffer([]byte{})
	confusionLen, _ := rand.Int(rand.Reader, big.NewInt(64))
	log.Println("sent confusion len is", uint8(confusionLen.Uint64()))
	confusion := make([]byte, int(confusionLen.Int64()))
	n, err := rand.Read(confusion)
	if n != int(confusionLen.Int64()) || err != nil {
		return
	}

	timeDiff, _ := rand.Int(rand.Reader, big.NewInt(60))
	timestamp := time.Now().Unix() + timeDiff.Int64() - 30 // time +-30
	log.Println("sent timestamp is", timestamp, uint32(timestamp))

	hnp := strings.Split(address, ":")
	host := []byte(hnp[0])
	port, _ := strconv.Atoi(hnp[1])

	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(confusionLen.Uint64()))
	_ = binary.Write(bytesBuffer, binary.BigEndian, confusion)
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint32(timestamp))
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(1))         // directive
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(len(host))) // directive
	_ = binary.Write(bytesBuffer, binary.BigEndian, host)
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint16(port))
	_ = binary.Write(bytesBuffer, binary.BigEndian, payload)

	body := bytesBuffer.Bytes()

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	digest := h.Sum(nil)
	log.Println("length of digest is", len(digest))

	return append(digest, body...), nil
}

func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}
