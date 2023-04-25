package protocols

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type BTPRequest struct {
	Address      string
	Payload      []byte
	digest       string
	confusionLen int
	timediff     int
	headerLen    int
	rawdata      []byte
}

func (request *BTPRequest) Validate(secret string) (err error) {
	if request.confusionLen < 0 || request.confusionLen > 64 {
		return errors.New("unexpected confusion length")
	}
	if request.timediff < -210 || 210 < request.timediff {
		return errors.New("timeout or replay attack")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(request.rawdata[32:])
	if hex.EncodeToString(h.Sum(nil)) != request.digest {
		return errors.New("digest mismatch, unauthorized connect")
	}

	lru = GetBtpCache()
	err = lru.Add(request.digest)
	if err != nil {
		return err
	}

	if request.headerLen != 41 {
		return errors.New("wrong btp head len")
	}

	return nil
}

func ParseBtpRequest(buffer []byte) (request *BTPRequest, err error) {
	request = &BTPRequest{}
	request.digest = hex.EncodeToString(buffer[:32])
	request.confusionLen = int(buffer[32]) // uint8

	pos := 32 + 1 + int(request.confusionLen)
	timestamp := binary.BigEndian.Uint32(buffer[pos : pos+4])
	request.timediff = int(time.Now().Unix()) - int(timestamp)

	pos += 5 // 4 for timestamp 1 for directives
	hostLen := int(buffer[pos])
	pos++
	host := string(buffer[pos : pos+hostLen])
	pos += hostLen
	port := strconv.Itoa(int(binary.BigEndian.Uint16(buffer[pos : pos+2])))
	pos += 2

	request.headerLen = pos - request.confusionLen - hostLen
	request.Address = host + ":" + port
	request.Payload = buffer[pos:]
	request.rawdata = buffer
	return request, nil
}

func EncodeBtpRequest(address string, payload []byte, secret string) (res []byte, err error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	confusionLen, _ := rand.Int(rand.Reader, big.NewInt(64))
	confusion := make([]byte, int(confusionLen.Int64()))
	n, err := rand.Read(confusion)
	if n != int(confusionLen.Int64()) || err != nil {
		return
	}

	timeDiff, _ := rand.Int(rand.Reader, big.NewInt(60))
	timestamp := time.Now().Unix() + timeDiff.Int64() - 30 // time +-30

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

	return append(digest, body...), nil
}
