package savior

import (
	"github.com/mylilcat/savior/launcher"
	"github.com/mylilcat/savior/net/kcp/connection"
	"github.com/mylilcat/savior/service"
	"os"
	"os/signal"
)

func Start(services ...*service.Service) {
	for _, s := range services {
		service.Register(s)
	}
	service.ServicesRun()
	launcher.ServerStart()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	service.ServicesStop()
}

func BindPort(port string) {
	launcher.SetPort(port)
}

func KcpOnConnectHandler(fun func(connection *connection.KCPConnection)) {
	launcher.SetKcpConnectHandler(fun)
}

func KcpOnDisconnectHandler(fun func(connection *connection.KCPConnection)) {
	launcher.SetKcpDisconnectHandler(fun)
}

func KcpOnReadHandler(fun func(connection *connection.KCPConnection, data []byte)) {
	launcher.SetKcpReadHandler(fun)
}

func KcpOnIdleHandler(fun func(connection *connection.KCPConnection)) {
	launcher.SetKcpIdleHandler(fun)
}
