package proxy

// implicit proxy interface definition
type Proxy interface {
	initializeProxy() error
	waitClient() error
	connServer() error
	showCommunication() error
	closeProxy() error
}
