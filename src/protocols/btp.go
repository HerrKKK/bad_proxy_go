package protocols

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"go_proxy/structure"
	"math/big"
	"strconv"
	"strings"
	"sync"
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

var btpLRU *structure.LRU[string]
var once sync.Once

func GetBtpCache() *structure.LRU[string] {
	once.Do(func() {
		instance := &structure.LRU[string]{}
		instance.Init(210000)
		btpLRU = instance
	})
	return btpLRU
}

func (request *BTPRequest) Validate(secret string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("btp validation panic")
		}
	}()
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

	lru := GetBtpCache()
	err = lru.Add(request.digest)
	if err != nil {
		return err // possible replay attack
	}

	if request.headerLen != btpHeaderLen {
		return errors.New("wrong btp head len")
	}

	return nil
}

func ParseBtpRequest(rawdata []byte) (request *BTPRequest, err error) {
	defer func() {
		if e := recover(); e != nil {
			request.Payload = rawdata // restore raw data
			err = errors.New("btp parse panic")
		}
	}()
	request = &BTPRequest{}
	request.digest = hex.EncodeToString(rawdata[:btpDigestLen])
	request.confusionLen = int(rawdata[btpDigestLen]) // uint8

	pos := btpDigestLen + btpConfusionLenDig + request.confusionLen
	timestamp := binary.BigEndian.Uint32(rawdata[pos : pos+btpTimestampLen])
	request.timediff = int(time.Now().Unix()) - int(timestamp)

	pos += btpTimestampLen + btpDirectiveDig
	hostLen := int(rawdata[pos])
	pos += btpHostLenDig
	host := string(rawdata[pos : pos+hostLen])
	pos += hostLen // possible out of bound
	port := strconv.Itoa(int(binary.BigEndian.Uint16(rawdata[pos : pos+btpPortLen])))
	pos += btpPortLen

	request.headerLen = pos - request.confusionLen - hostLen
	request.Address = host + ":" + port
	request.Payload = rawdata[pos:]
	request.rawdata = rawdata
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
