package protocols

import (
	"C"
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
	Url     string
	Address string
	Payload []byte
}

func ParseHttpRequest(buffer []byte) (request HTTPRequest, err error) {
	if _, err = fmt.Sscanf(
		string(buffer[:bytes.IndexByte(buffer[:], '\n')]),
		"%s%s",
		&request.Method,
		&request.Url,
	); err != nil {
		return
	}
	hostPortURL, err := url.Parse(request.Url)
	if err != nil { // Just believe url is an ip when parsing failed, leave err behind.
		request.Address = request.Url
	} else if len(hostPortURL.Host) == 0 { // for opaque urls.
		request.Address = hostPortURL.Scheme + ":" + hostPortURL.Opaque
	} else { // url parsed successfully.
		request.Address = hostPortURL.Host
	}

	if strings.Index(request.Address, ":") == -1 {
		request.Address = request.Address + ":80"
	}

	request.Payload = buffer[bytes.Index(buffer[:], []byte("\r\n\r\n")):]
	return
}

type HttpInbound struct {
	Conn net.Conn
}

func (inbound *HttpInbound) Fallback(rawData []byte) {
	_ = rawData
}

func (inbound *HttpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196)
	length, err := inbound.Conn.Read(payload[:])
	if err != nil {
		return
	}
	request, err := ParseHttpRequest(payload[:])
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
