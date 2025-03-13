package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type WSConnection struct {
	id              any
	conn            *websocket.Conn
	ioWorker        *worker
	isConnected     bool
	closeNotifyChan chan *WSConnection
	messageType     int
	buffer          []byte
}

func NewWSConnection(conn *websocket.Conn, closeNotifyChan chan *WSConnection, handler *Handler, messageType int) *WSConnection {
	ws := new(WSConnection)
	ws.conn = conn
	ws.closeNotifyChan = closeNotifyChan
	ws.isConnected = true
	ws.messageType = messageType
	ws.ioWorker = newIOWorker(ws, "ws", handler)
	return ws
}

func (ws *WSConnection) GetId() any {
	return ws.id
}

func (ws *WSConnection) SetId(id any) {
	ws.id = id
}

func (ws *WSConnection) IsConnected() bool {
	return ws.isConnected
}

func (ws *WSConnection) Close() {
	ws.isConnected = false
	ws.conn.Close()
	ws.closeNotifyChan <- ws
}

func (ws *WSConnection) Read(b []byte) (n int, err error) {
	if len(ws.buffer) > 0 {
		n = copy(b, ws.buffer)
		ws.buffer = ws.buffer[n:]
		return n, nil
	}
	messageType, msgBytes, err := ws.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	if messageType != websocket.BinaryMessage && messageType != websocket.TextMessage {
		return 0, fmt.Errorf("mssage type not supported: %d", messageType)
	}

	n = copy(b, msgBytes)
	if n < len(msgBytes) {
		ws.buffer = msgBytes[n:]
	}
	return n, nil
}

func (ws *WSConnection) Write(b []byte) (n int, err error) {
	if ws.messageType == websocket.TextMessage {
		err = ws.conn.WriteMessage(websocket.TextMessage, b)
		if err != nil {
			return 0, err
		}
	}
	if ws.messageType == websocket.BinaryMessage {
		err = ws.conn.WriteMessage(websocket.BinaryMessage, b)
		if err != nil {
			return 0, err
		}
	}
	return len(b), nil
}

func (ws *WSConnection) Send(b []byte) {
	if ws.isConnected {
		ws.ioWorker.sender.send(b)
	}
}

func (ws *WSConnection) GetLastReadTime() time.Time {
	return ws.ioWorker.receiver.lastReadTime
}

func (ws *WSConnection) GetLastWriteTime() time.Time {
	return ws.ioWorker.sender.lastWriteTime
}

func (ws *WSConnection) GetRemoteAddr() net.Addr {
	return ws.conn.RemoteAddr()
}
