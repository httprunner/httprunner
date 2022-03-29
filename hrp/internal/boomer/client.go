package boomer

type client interface {
	connect() (err error)
	close()
	recvChannel() chan *genericMessage
	sendChannel() chan *genericMessage
	disconnectedChannel() chan bool
}
