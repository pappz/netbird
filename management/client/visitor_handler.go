package client

type VisitorChannel interface {
	Send(session string, data []byte) error
}

type VisitorHandler interface {
	// SetChannel todo: move to another layer
	SetChannel(VisitorChannel)

	OnNewConnection(sessionId string) error
	OnNewMsg(sessionId string, msg []byte) error
	OnCloseConnection(sessionId string) error
}
