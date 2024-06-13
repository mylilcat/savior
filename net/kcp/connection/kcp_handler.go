package connection

var OnConnect func(connection *KCPConnection)

var OnDisconnect func(connection *KCPConnection)

var OnRead func(connection *KCPConnection, data []byte)

var OnIdle func(connection *KCPConnection)
