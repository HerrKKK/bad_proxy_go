package protocols

type BTPRequest struct {
	Address string
	Payload []byte
}

func (request BTPRequest) Parse(buffer []byte) (BTPRequest, error) {
	addrLen := buffer[0] // uint8
	//fmt.Println("btp addr is")
	//fmt.Println(string(buffer[1 : 1+addrLen]))
	request.Address = string(buffer[1 : 1+addrLen])
	request.Payload = buffer[1+addrLen:]
	return request, nil
}
