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

const btpDigestLen = 32
const btpConfusionLenDig = 1
const btpTimestampLen = 4
const btpDirectiveDig = 1
const btpHostLenDig = 1
const btpPortLen = 2
const btpHeaderLen = btpDigestLen +
	btpConfusionLenDig +
	btpTimestampLen +
	btpHostLenDig +
	btpDirectiveDig +
	btpPortLen
const timeThreshold = 210
const btpMaxConfusionLen = 64
const btpTimeDiffRand = 30

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
	if request.confusionLen < 0 || request.confusionLen > btpMaxConfusionLen {
		return errors.New("unexpected confusion length")
	}
	if request.timediff < -timeThreshold || timeThreshold < request.timediff {
		return errors.New("timeout or replay attack")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(request.rawdata[btpDigestLen:])
	if hex.EncodeToString(h.Sum(nil)) != request.digest {
		return errors.New("digest mismatch, unauthorized connect")
	}

	lru = GetBtpCache()
	err = lru.Add(request.digest)
	if err != nil {
		return err
	}

	if request.headerLen != btpHeaderLen {
		return errors.New("wrong btp head len")
	}

	return nil
}

func ParseBtpRequest(buffer []byte) (request *BTPRequest, err error) {
	request = &BTPRequest{}
	request.digest = hex.EncodeToString(buffer[:btpDigestLen])
	request.confusionLen = int(buffer[btpDigestLen]) // uint8

	pos := btpDigestLen + btpConfusionLenDig + request.confusionLen
	timestamp := binary.BigEndian.Uint32(buffer[pos : pos+btpTimestampLen])
	request.timediff = int(time.Now().Unix()) - int(timestamp)

	pos += btpTimestampLen + btpDirectiveDig
	hostLen := int(buffer[pos])
	pos += btpHostLenDig
	host := string(buffer[pos : pos+hostLen])
	pos += hostLen // possible out of bound
	port := strconv.Itoa(int(binary.BigEndian.Uint16(buffer[pos : pos+btpPortLen])))
	pos += btpPortLen

	request.headerLen = pos - request.confusionLen - hostLen
	request.Address = host + ":" + port
	request.Payload = buffer[pos:]
	request.rawdata = buffer
	return request, nil
}

func EncodeBtpRequest(address string, payload []byte, secret string) (res []byte, err error) {
	bytesBuffer := bytes.NewBuffer([]byte{})
	confusionLen, err := rand.Int(rand.Reader, big.NewInt(btpMaxConfusionLen))
	if err != nil {
		return
	}
	confusion := make([]byte, int(confusionLen.Int64()))
	n, err := rand.Read(confusion)
	if n != int(confusionLen.Int64()) || err != nil {
		return
	}

	timeDiff, err := rand.Int(rand.Reader, big.NewInt(2*btpTimeDiffRand))
	if err != nil {
		return
	}
	timestamp := time.Now().Unix() + timeDiff.Int64() - btpTimeDiffRand // time +-30

	hnp := strings.Split(address, ":")
	host := []byte(hnp[0])
	port, err := strconv.Atoi(hnp[1])
	if err != nil {
		return
	}

	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(confusionLen.Uint64()))
	_ = binary.Write(bytesBuffer, binary.BigEndian, confusion)
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint32(timestamp))
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(1)) // directive
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint8(len(host)))
	_ = binary.Write(bytesBuffer, binary.BigEndian, host)
	_ = binary.Write(bytesBuffer, binary.BigEndian, uint16(port))
	_ = binary.Write(bytesBuffer, binary.BigEndian, payload)

	body := bytesBuffer.Bytes()

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	digest := h.Sum(nil)
	res = append(digest, body...)
	return
}
