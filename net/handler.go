package net

type Handler struct {
	OnConnect    func(c Connection)
	OnDisconnect func(c Connection)
	OnRead       func(c Connection, data []byte)
	OnIdle       func(c Connection)
}
