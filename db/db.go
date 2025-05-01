package db

type DB interface {
	UpdateState(channel string, callback func(state *ChannelState))
	GetState(channel string) ChannelState
	GetAllStates() []ChannelState
	GetAllChannelNames() []string
	ReadAllStates(callback func(state *ChannelState))
}
