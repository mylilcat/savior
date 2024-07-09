package savior

import (
	"github.com/mylilcat/savior/launcher"
	"github.com/mylilcat/savior/net"
	"github.com/mylilcat/savior/service"
	"os"
	"os/signal"
	"time"
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

func SetProto(p int) {
	launcher.SetProto(p)
}

func SetOnConnectHandler(f func(c net.Connection)) {
	net.OnConnect = f
}

func SetOnDisconnectHandler(f func(c net.Connection)) {
	net.OnDisconnect = f
}

func SetOnReadHandler(f func(c net.Connection, data []byte)) {
	net.OnRead = f
}

func SetOnIdleHandler(f func(c net.Connection)) {
	net.OnIdle = f
}

func SetIdleMonitor(readIdle int64, writeIdle int64, unit time.Duration) {
	launcher.SetIdleMonitor(readIdle, writeIdle, unit)
}
