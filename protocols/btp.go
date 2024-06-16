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
	"io"
	"log"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	btpDigestLen       = 32
	btpConfusionLenDig = 1
	btpTimestampLen    = 4
	btpDirectiveDig    = 1
	btpHostLenDig      = 1
	btpPortLen         = 2
	btpHeaderLen       = btpDigestLen +
		btpConfusionLenDig +
		btpTimestampLen +
		btpHostLenDig +
		btpDirectiveDig +
		btpPortLen
	timeThreshold      = 210
	btpMaxConfusionLen = 64
	btpTimeDiffRand    = 30
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

func (request *BTPRequest) validate(secret string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
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

	err = GetBtpCache().Add(request.digest)
	if err != nil {
		return err // possible replay attack
	}

	if request.headerLen != btpHeaderLen {
		return errors.New("wrong btp head len")
	}

	return nil
}

func parseBtpRequest(rawdata []byte) (request *BTPRequest, err error) {
	defer func() {
		if r := recover(); r != nil {
			request.Payload = rawdata // restore raw data
			log.Println(r)
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

func encodeBtpRequest(address string, payload []byte, secret string) (res []byte, err error) {
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

type BtpInbound struct {
	Conn   net.Conn
	Secret string
}

func (inbound *BtpInbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	out, err := net.Dial("tcp", reverseLocalAddr)
	defer inbound.Close()
	defer out.Close()
	if err != nil {
		return
	}
	_, _ = out.Write(rawdata)

	go io.Copy(inbound, out)
	_, _ = io.Copy(out, inbound)
	return
}

func (inbound *BtpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196) // return rawdata on error
	length, err := inbound.Conn.Read(payload)
	if err != nil {
		return
	}
	request, err := parseBtpRequest(payload[:length])
	if err != nil {
		return
	}
	err = request.validate(inbound.Secret)
	if err != nil { // try to handle http connection
		return
	}
	return request.Address, request.Payload, nil
}

func (inbound *BtpInbound) Read(b []byte) (int, error) {
	return inbound.Conn.Read(b)
}

func (inbound *BtpInbound) Write(b []byte) (int, error) {
	return inbound.Conn.Write(b)
}

func (inbound *BtpInbound) Close() error {
	return inbound.Conn.Close()
}

type BtpOutbound struct {
	Conn   net.Conn
	Secret string
}

func (outbound *BtpOutbound) Connect(targetAddr string, payload []byte) (err error) {
	payload, err = encodeBtpRequest(targetAddr, payload, outbound.Secret)
	if err != nil {
		return
	}
	_, err = outbound.Conn.Write(payload)
	return
}

func (outbound *BtpOutbound) Read(b []byte) (int, error) {
	return outbound.Conn.Read(b)
}

func (outbound *BtpOutbound) Write(b []byte) (int, error) {
	return outbound.Conn.Write(b)
}

func (outbound *BtpOutbound) Close() error {
	return outbound.Conn.Close()
}
