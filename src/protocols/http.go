package protocols

import (
	"bytes"
	"errors"
	"fmt"
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

func (request HTTPRequest) Parse(buffer []byte) (HTTPRequest, error) {
	_, err := fmt.Sscanf(
		string(buffer[:bytes.IndexByte(buffer[:], '\n')]),
		"%s%s",
		&request.Method,
		&request.Url,
	)
	if err != nil {
		return HTTPRequest{}, errors.New("scan error")
	}
	hostPortURL, _ := url.Parse(request.Url)
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
