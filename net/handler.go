package net

// OnConnect connection accept handler.
// 新的连接接入处理函数
var OnConnect func(c Connection)

// OnDisconnect connection disconnect handler.
// 连接断开处理
var OnDisconnect func(c Connection)

// OnRead Message handler.
// 读到消息的处理
var OnRead func(c Connection, data []byte)

// OnIdle idle handler.
// 空闲连接处理
var OnIdle func(c Connection)

type Handler struct {
	OnConnect    func(c Connection)
	OnDisconnect func(c Connection)
	OnRead       func(c Connection, data []byte)
	OnIdle       func(c Connection)
}

//func NewHandler() *Handler {
//	h := &Handler{
//		OnConnect:    OnConnect,
//		OnDisconnect: OnDisconnect,
//		OnRead:       OnRead,
//		OnIdle:       OnIdle,
//	}
//	OnConnect = nil
//	OnDisconnect = nil
//	OnRead = nil
//	OnIdle = nil
//	return h
//}
