package main

import (
	"C"
	"go_proxy/protocols"
)

//export ParseHttpRequest
func ParseHttpRequest(rawData []byte) (method string, url string, address string, body []byte, err error) {
	httpRequest, err := protocols.ParseHttpRequest(rawData)
	return httpRequest.Method, httpRequest.Url, httpRequest.Address, httpRequest.Payload, err
}
