package transport

type Transport interface {
	Read(buffer []byte) (int, error)
	Write(buffer []byte) (int, error)
	Close() error
}

type Network interface {
	Accept() (Transport, error)
	Dial(address string) (Transport, error)
	Listen(address string) (Network, error)
}
