package transport

type Transport interface {
	Start(addr string) error
}
