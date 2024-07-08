package savior

import (
	"github.com/mylilcat/savior/launcher"
	"github.com/mylilcat/savior/net"
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

func SetOnConnectHandler(f func(c net.Connection)) {
	net.OnConnect = f
}

func SetOnDisconnectHandler() {

}
