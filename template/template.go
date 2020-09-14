package template

// implicit proxy interface definition
type Proxy interface {
	ReadyToCommunicate() error
	CloseProxy()
}
