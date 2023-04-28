package protocols

import (
	"bytes"
	"fmt"
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
