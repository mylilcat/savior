package net

import "sync"

var OnConnect func(c Connection)

var OnDisconnect func(c Connection)

var OnRead func(c Connection, data []byte)

var OnIdle func(c Connection)

var IdleMonitoring func(connections *sync.Map)
