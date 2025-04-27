package producer

type Producer interface {
	Run() error
	Stop()
	AddChannel(channel string)
	RemoveChannel(channel string)
}
