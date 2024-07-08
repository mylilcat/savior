package net

type Connection interface {
	GetId() any
	SetId(id any)
	IsConnected() bool
	Close()
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
}
