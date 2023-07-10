package protocols

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type Request interface {
	Parse()
}

type HTTPRequest struct {
	Method  string
	url     string
	Address string
	Payload []byte
}

func (request HTTPRequest) Parse(buffer []byte) (req HTTPRequest, err error) {
	_, err = fmt.Sscanf(
		string(buffer[:bytes.IndexByte(buffer[:], '\n')]),
		"%s%s",
		&request.Method,
		&request.url,
	)
	if err != nil {
		return
	}
	hostPortURL, err := url.Parse(request.url)
	if err != nil {
		return
	}
	if len(hostPortURL.Host) == 0 {
		request.Address = hostPortURL.Scheme + ":" + hostPortURL.Opaque
	} else {
		request.Address = hostPortURL.Host
	}
	if strings.Index(request.Address, ":") == -1 {
		request.Address = request.Address + ":80"
	}

	request.Payload = buffer[bytes.Index(buffer[:], []byte("\r\n\r\n")):]
	return request, nil
}

type HttpInbound struct {
	Conn net.Conn
}

func (inbound *HttpInbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	_ = reverseLocalAddr
	_ = rawdata
	return
}

func (inbound *HttpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196)
	length, err := inbound.Conn.Read(payload[:])
	if err != nil {
		return
	}
	request, err := HTTPRequest{}.Parse(payload[:])
	if err != nil {
		return
	}
	targetAddr = request.Address
	if request.Method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		_, err = inbound.Conn.Write([]byte(response))
		if err != nil {
			return
		}
		payload = make([]byte, 8196) // clear
		length, err = inbound.Conn.Read(payload)
		if err != nil {
			return
		}
	}
	payload = payload[:length]
	return
}

func (inbound *HttpInbound) Read(b []byte) (int, error) {
	return inbound.Conn.Read(b)
}

func (inbound *HttpInbound) Write(b []byte) (int, error) {
	return inbound.Conn.Write(b)
}

func (inbound *HttpInbound) Close() error {
	return inbound.Conn.Close()
}
