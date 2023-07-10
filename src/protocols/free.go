package protocols

import "net"

type FreeOutbound struct {
	Conn net.Conn
}

func (outbound *FreeOutbound) Connect(targetAddr string, payload []byte) (err error) {
	_ = targetAddr
	_, err = outbound.Conn.Write(payload)
	return
}

func (outbound *FreeOutbound) Read(b []byte) (int, error) {
	return outbound.Conn.Read(b)
}

func (outbound *FreeOutbound) Write(b []byte) (int, error) {
	return outbound.Conn.Write(b)
}

func (outbound *FreeOutbound) Close() error {
	if outbound.Conn == nil {
		return nil
	}
	return outbound.Conn.Close()
}
