package producer

type Producer interface {
	Run() error
	Shutdown()
	AddChannel(channel string)
	RemoveChannel(channel string)
}
