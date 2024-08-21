Savior 轻量级游戏服务器框架
===========================
名字叫希维尔就是觉得很帅。和lol里的轮子妈没关系的。   
搞这个框架就是觉得，如果我想要快速搭建一个轻量级的socket server,以现有的框架来说都太重度，定制化程度略高。缺少了相对的自由度。
所以savior的设计的目标就是轻便快捷，使用友好，用最少的代码快速搭建长连接服务，后续还会谨慎的填充游戏服务相关的实用功能。   
目前支持TCP，KCP，以后会加上websocket。  
主打一手拒绝臃肿，提供选择，减少强制。

如何快速启动服务
---------------------------

```go
package main

import (
	"github.com/mylilcat/savior"
	"github.com/mylilcat/savior/net"
	"time"
)

func OnConnect(conn net.Connection) {
	//连接接入处理
}

func OnDisConnect(conn net.Connection) {
	//连接断开处理
}

func OnRead(conn net.Connection, data []byte) {
	//接收数据处理
}

func OnIdle(conn net.Connection) {
	//连接空闲处理
}

func main() {
	savior.BindPort("8000") //设置服务端口
	savior.SetProto("tcp") //设置协议 使用kcp协议填 "kcp" 
	savior.SetOnConnectHandler(OnConnect)
	savior.SetOnDisconnectHandler(OnDisConnect)
	savior.SetOnReadHandler(OnRead)
	savior.SetOnIdleHandler(OnIdle)
	savior.SetIdleMonitor(5, 5, time.Second) //设置空闲检测 
	savior.Start() //启动
}

```
以上就是启动一个socket server的全部流程。下面介绍savior提供的一些较为实用的功能。

连接空闲检测
---------------------------
```go
func SetIdleMonitor(readIdle int64, writeIdle int64, unit time.Duration)
```

* readIdle -> read超时时间。
* writeIdle -> write超时时间。
* unit -> 超时时间单位，最小支持到毫秒。
* 注意：最小超时时间为15毫秒。

如下，设置读写空闲检测5秒，连接超过5秒没有读行为或者超过5秒没有写行为，会触发OnIdle方法。如果不需要空闲检测，不设置IdleMonitor即可。
```go
savior.SetIdleMonitor(5, 5, time.Second)
```

Connection如何融入到（游戏）业务中
---------------------------
首先，每个设置过的handler方法，入参都有当前对应的Connection，触发不同的handler你都能得到当前的连接，Connection接口中有GetId()方法和SetId()方法，这个ID就可以用来标识连接连接归属。  
思考了很久，还是觉得online的管理应该是交由使用者去处理,savior能提供最简单的辅助就是给conn添加所属标志。

```go
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

type player struct {
	uid  string
	conn Connection
}

// 举例，如接收到玩家登录消息
func OnRead(conn net.Connection, data []byte) {
  id := string(data)
  conn.SetId(id)
  //do login
  p := new(player)
  p.uid = id
  p.conn = conn
  //add to online manager
  ...
}
```
时间轮定时器使用
------------------------
目前定时任务分为两种，延时任务 DelayTask，和轮询任务 IntervalTask。

创建时间轮定时器
```go
// NewTimer Initialize a time wheel timer. 初始化时间轮定时器
// time unit,supports down to milliseconds. 时间间隔单位，最小到毫秒。
// slot,number of slots in the time wheel. 时间轮的槽位数量
// period,attention pls,if you set unit milliseconds,the minimum time interval is 15 milliseconds!!!!!!!!!!!!!!
// period，是定时器时间间隔，如果你设置的时间单位是毫秒，那period最小支持到15毫秒。
func NewTimer(period int64, unit time.Duration, slotNum int) *Timer
```

* period -> 任务执行时间间隔，最小15毫秒。
* unit -> 时间间隔单位，最小支持毫秒。
* slotNum -> 时间轮槽位








