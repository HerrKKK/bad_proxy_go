package protocols

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type HTTPRequest struct {
	method  string
	url     string
	path    string
	address string
	payload []byte
}

func parseHttpRequest(buffer []byte) (request HTTPRequest, err error) {
	_, err = fmt.Sscanf(
		string(buffer[:bytes.IndexByte(buffer, '\n')]),
		"%s%s",
		&request.method,
		&request.url,
	)
	if err != nil {
		return
	}
	requestUrl, err := url.Parse(request.url)
	if err != nil { // Just believe url is an ip when parsing failed, leave err behind.
		request.address = request.url
	} else if len(requestUrl.Host) == 0 { // for opaque urls.
		request.address = requestUrl.Scheme + ":" + requestUrl.Opaque
	} else { // url parsed successfully.
		request.address = requestUrl.Host
	}

	request.path = requestUrl.Path
	if strings.Index(request.address, ":") == -1 {
		request.address = request.address + ":80"
	}

	request.payload = buffer[bytes.Index(buffer, []byte("\r\n\r\n")):]
	return request, nil
}

type HttpInbound struct {
	Conn net.Conn
}

func (inbound *HttpInbound) Fallback(reverseLocalAddr string, rawdata []byte) {
	_, _ = reverseLocalAddr, rawdata
}

func (inbound *HttpInbound) Connect() (targetAddr string, payload []byte, err error) {
	payload = make([]byte, 8196)
	length, err := inbound.Read(payload)
	if err != nil {
		return
	}
	request, err := parseHttpRequest(payload)
	if err != nil {
		return
	}
	targetAddr = request.address
	if request.method == "CONNECT" {
		var response = "HTTP/1.1 200 Connection Established\r\nConnection: close\r\n\r\n"
		_, err = inbound.Write([]byte(response))
		if err != nil {
			return
		}
		payload = make([]byte, 8196) // clear
		length, err = inbound.Read(payload)
		if err != nil {
			return
		}
	}
	payload = payload[:length]
	/* replace url with path */
	payload = bytes.Replace(payload, []byte(request.url), []byte(request.path), 1)
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
