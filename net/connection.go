package net

import "time"

type Connection interface {
	// GetId get the ID based on your logic,for example the player ID.
	// 获得连接的业务ID，如果是游戏服务器，这个ID就可以是玩家的ID。
	GetId() any
	// SetId set the ID based on your logic,for example the player ID.
	// 设置连接的业务ID，比如玩家登录过后，把连接ID设置成玩家的ID。
	SetId(id any)
	IsConnected() bool
	Close()
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Send(b []byte)
	GetLastReadTime() time.Time
	GetLastWriteTime() time.Time
}
