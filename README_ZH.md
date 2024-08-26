Savior golang轻量级游戏服务器框架
===========================
名字叫希维尔就是觉得很帅。  
搞这个框架就是觉得，如果我想要快速搭建一个轻量级的socket server,以现有的框架来说都太重度，定制化程度略高。缺少了相对的自由度。
所以savior的设计的目标就是轻便快捷，使用友好，用最少的代码能够快速搭建长连接服务。   
目前支持TCP，KCP，以后会加上websocket。  
主打一手拒绝臃肿，提供选择，减少强制。
当然后续还会谨慎的填充游戏服务相关的实用功能。该不该实现很多还没想好，很纠结。 比如消息路由，我觉得暴露给使用者的，到字节就够了。数据的codec理应是使用者自由发挥的地方。手动编解码也好，或者protobuf，都应该属于高自由度的功能。

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
以上就是savior启动一个socket server的全部流程。下面介绍savior提供的一些较为实用的功能。

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
首先，每个设置过的handler方法，入参都有当前对应的Connection，触发不同的handler你都能得到当前的连接，Connection接口中有GetId()方法和SetId()方法，这个ID就可以用来标识连接的归属。  
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
    p.conn.Write([]byte("登录成功"))
}
```
时间轮定时器使用
------------------------
时间轮就不用多说了，相比于全遍历和最小堆顶来说是性价比最高的定时器实现方式。   
目前支持的定时任务分为两种，延时任务 DelayTask，和轮询任务 IntervalTask。

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
* slotNum -> 时间轮槽位，槽位越多，执行精度越细。

添加任务
```go
// AddTask add task to the scheduler. 添加延时任务
// f task
// delayTime execute task after how long a delay,the delay time unit is the one you set when initializing the timer.
// delayTime 延时的时间，如果你初始化定时器时设置的时间单位是秒，delayTime = 2，那就是2秒后执行。如果设置的是分钟，那就是2分钟后执行。
// typ task type are divided into delay tasks (DelayTask) and interval tasks (IntervalTask). If no value is provided, the default is a delay task.
// typ 任务类型分为两种，(DelayTask)延时任务，以及轮询任务(IntervalTask)，不传值默认为延时任务。
func (t *Timer) AddTask(f func(), delayTime int64, typ ...any)
```

* f 要执行的任务
* delayTime 延时时长，假如delayTime = 5，如果创建定时器时设置的时间单位是秒，就是5秒后执行。
* typ 任务类型，不传值默认为延时任务，传1，（timer.IntervalTask），是轮询执行，同上，delayTime = 5，就是每隔5秒执行一次。

启动和关闭
```go
func (t *Timer) Start()
func (t *Timer) Stop()
```
使用示例
```go
package main

import (
    "github.com/mylilcat/savior/timer"
    "time"
)

func main() {
    //创建一个执行间隔为1秒，时间轮槽位为128的定时器
    taskTimer := timer.NewTimer(1, time.Second, 128)
    taskTimer.Start()
    //添加5秒后执行的延时任务
    taskTimer.AddTask(func() {
        //do something
    }, 5)
    
    //添加每隔5秒执行一次的定时任务，首次执行也是5秒后。
    taskTimer.AddTask(func() {
        //do something
    }, 5, timer.IntervalTask)
}

```
有限状态机的使用
--------------------
savior当前应用在的项目是一款轻度类MMORPG的游戏上，游戏中要用状态机实现的功能有控制boss和npc的行为，包括释放技能，寻敌，警戒，和其他的一些逻辑行为。   
状态不是很多，感觉用不上行为树，都是些简单的状态交替。所以就搞了一个小型的fsm。目前支持的状态转换条件类型只有顺序判断（按顺序check转换条件）。后续会考虑增加和完善，比如随机状态转换。

状态机的使用实例
```go
package main

import (
    "github.com/mylilcat/savior/fsm"
    "math/rand"
    "os"
    "os/signal"
    "time"
)

type boss struct {
    name string
    fsm  *fsm.FiniteStateMachine
}

func main() {
    boss := new(boss)
    boss.name = "final boss"
    
    //寻敌状态
    enemySearchingState := fsm.NewState("enemySearching")
    //攻击状态
    attackState := fsm.NewState("attack")
    
    //初始化状态为寻敌
    boss.fsm = fsm.NewFiniteStateMachine(enemySearchingState)
    //boss 进入到寻敌状态要执行的方法
    enemySearchingState.SetOnEnter(func() {
        println(boss.name + " enter the state of searching for enemies")
    })
    //boss 退出寻敌状态要执行的方法
    enemySearchingState.SetOnExist(func() {
        println(boss.name + " leave the state of searching for enemies")
    })
    //boss 正在寻找敌人
    enemySearchingState.SetOnExecute(func() {
        println(boss.name + " currently searching for enemies")
    })
    
    // 转换到攻击状态的条件
    toAttackTransition := fsm.NewTransition(attackState)
    toAttackTransition.SetConvertCondition(func() bool {
        //模拟寻敌
        if rand.Intn(100) > 50 {
            //找到敌人，进入到攻击状态
            return true
        }
        return false
    })
    // 加入到寻敌状态转换列表
    enemySearchingState.AddTransitions(toAttackTransition)
    
    //boss 进入到攻击状态要执行的方法
    attackState.SetOnEnter(func() {
        println(boss.name + " enter the state of attack")
    })
    
    //boss 退出攻击状态要执行的方法
    attackState.SetOnExist(func() {
        println(boss.name + " leave the state of attack")
    })
    
    //boss 正在攻击
    attackState.SetOnExecute(func() {
        println(boss.name + " attacking")
    })
    
    //准换到寻敌状态的条件
    toSearchingState := fsm.NewTransition(enemySearchingState)
    toSearchingState.SetConvertCondition(func() bool {
        //模拟攻击
        if rand.Intn(100) > 50 {
            //攻击结束，进入到寻敌状态
            return true
        }
        return false
    })
    //加入到攻击状态转换列表
    attackState.AddTransitions(toSearchingState)
    //设置执行间隔为50毫秒
    boss.fsm.SetPeriodAndUnit(50, time.Millisecond)
    boss.fsm.Start()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Kill)
    <-sigChan
    boss.fsm.Stop()
}

```
最终呈现的结果就是boss在寻敌和攻击两种状态之间切换，寻找到目标就进入到攻击状态，攻击结束进入到寻敌状态。  

部分日志如下：
```
final boss enter the state of searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss leave the state of searching for enemies
final boss enter the state of attack
final boss attacking
final boss attacking
final boss attacking
final boss leave the state of attack
final boss enter the state of searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss currently searching for enemies
final boss leave the state of searching for enemies
final boss enter the state of attack
final boss attacking
final boss attacking
final boss attacking
final boss attacking
final boss leave the state of attack
final boss enter the state of searching for enemies
final boss currently searching for enemies
final boss leave the state of searching for enemies
final boss enter the state of attack
```
当然了你也可以灵活的运用这个工具，比如简单的单一执行状态，可以只设置状态机的执行函数，自己管理状态机start和stop的时机。   
比如boss的技能是发射火球，服务器全量去跑火球的执行状态，就可以只设置火球的execute。

```go
package main

import (
    "github.com/mylilcat/savior/fsm"
    "math/rand"
    "os"
    "os/signal"
)

type fireBall struct {
    fsm *fsm.FiniteStateMachine
}

func main() {

    fireBall := new(fireBall)
    flyingState := fsm.NewState("flying")
    fireBall.fsm = fsm.NewFiniteStateMachine(flyingState)
    flyingState.SetOnExecute(func() {
        //模拟火球飞行中是否触碰到物体
        if rand.Intn(100) > 50 {
            // 触碰到物体，状态机停止
            fireBall.fsm.Stop()
        }
        // 没有触碰到物体，继续飞行
        println("flying")
    })
    fireBall.fsm.SetPeriodAndUnit(100,time.Millisecond)
    fireBall.fsm.Start()
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Kill)
    <-sigChan
    fireBall.fsm.Stop()
	
}

```

Service服务管理的使用
------------------------
如果你想让你的游戏业务模块隔离性强，更容易的管理和通信，那么你可以考虑使用savior service，参考skynet的设计思维，service提供了较为完备且易于理解的actor通信机制，并且提供易于维护的创建和销毁生命周期。
一个service目前会提供一个简易的协程池来接收跨模块的函数调用，具体为三个channel为100的任务队列，根据队列当前任务数，挑选一个最少任务数的协程执行任务。也就是说每个service有3个协程来执行actor调用。
目前这个版本就是写死的，协程数和channel容量都是固定的，因为未来我想给其改造成协程和channel是可根据服务压力进行无锁动态伸缩。
服务停止，所有service协程中的actor任务会执行完毕再退出。
还是强调一下，如果觉得savior service太过繁琐，或者没有直调方便，或者自己能管理好自己的业务模块，那么完全可以不用这个功能。
service完全就是为了让业务之间的隔离性更强，而且对于有skynet使用习惯的开发者，会十分顺手。多协程也能进一步提高资源利用。

savior service提供的方法
```go
type Service struct {
    name      string //服务名称
    manager   *Manager //服务管理器 后续新增service功能的话，都会在这里做管理。比如模块链路追踪，流量分析，埋点，service状态可视化。
    wgManager sync.WaitGroup
}

// NewService make a service with actor,service name must be unique.
// 初始化一个actor服务类，服务名称要保证唯一。
func NewService(name string) *Service {
    if name == "" {
    panic("service name empty")
    }
    s := new(Service)
    s.name = name
    s.manager = NewManager()
    return s
}

// RegisterActorFunction register a service actor call function.
// 注册服务actor调用函数
func (s *Service) RegisterActorFunction(name string, function any) {
    s.manager.actor.RegisterFunction(name, function)
}

// SetInitFunc set service initialization function; for example,
// if this is a leaderboard service, you can load leaderboard data during the startup phase.
// 设置服务初始化函数，比如这是一个排行榜服务，就可以在启动阶段加载排行榜数据。
func (s *Service) SetInitFunc(fun func()) {
    s.manager.serviceInitFunc = fun
}

// SetDestroyFunc set a service shutdown handler; for example, save the in-memory leaderboard data after stopping the service.
// 设置服务关闭后处理函数，比如停服后保存内存中的排行榜数据。
func (s *Service) SetDestroyFunc(fun func()) {
    s.manager.serviceDestroyFunc = fun
}
```

下面是service具体的实用示例，比如我要创建一个在线玩家管理的service。那么以包为模块边界，package为online，online下创建一个online service。
1. online service 设置 InitFunc ，比如加载一些活跃玩家的缓存。
2. online service 设置 DestroyFunc ， 服务器停服，落地玩家数据。
3. 注册一个提供给campaign service（以活动服务为例）调用的广播（actor）方法。 
4. 注册一个获得在线玩家的（actor）方法。


```go
package online

import (
    "github.com/mylilcat/savior/net"
    "github.com/mylilcat/savior/service"
)

var Service *service.Service         //玩家在线管理service
var onlinePlayers map[string]*Player //在线玩家

type Player struct {
    id   string //玩家ID
    Conn *net.TCPConnection
}

func inti() {
    Service = service.NewService("online")
    onlinePlayers = make(map[string]*Player)
    Service.SetInitFunc(loadCache) //loadCache 会在所有service启动完成后执行
    Service.SetDestroyFunc(savePlayerData) //savePlayerData 服务停服后，所有service执行完正在处理的actor任务后，执行该方法。
    Service.RegisterActorFunction("broadcast", broadcast) //注册广播方法为service的actor方法，提供给campaign service调用。
    Service.RegisterActorFunction("getPlayer",getOnlinePlayerById) //获得在线玩家
}

//广播 示例广播的消息是字符串
func broadcast(msg string) {
    for _, p := range onlinePlayers {
        p.Conn.Send([]byte(msg))
    }
}

func getOnlinePlayerById(id string) *Player {
	return onlinePlayers[id]
}

//加载玩家缓存
func loadCache(){
	
}

//落地玩家数据
func savePlayerData(){
	
}

```

示例 活动service调用online service的actor方法。

```go
package campaign

import (
    "github.com/mylilcat/savior/net"
    "github.com/mylilcat/savior/service"
    "github.com/mylilcat/savior/timer"
    "online"
    "time"
)

var Service *service.Service //活动service
var campaignTimer = timer.NewTimer(1, time.Second,128) //定义一个活动定时器

func inti() {
    Service = service.NewService("campaign")
    campaignTimer.AddTask(func() {
        service.Call("online","broadcast","活动开始了") //调用online service的broadcast方法，通知在线玩家活动开始。
    },120) //120秒后，通知所有玩家活动开始
}

// 模拟给指定的在线玩家发奖励，主要是示范下，有返回值的actor方法如何调用。
func sendOnlineRewardToPlayer(id string) { 
	p := service.Call1[*online.Player]("online","getPlayer",id) //泛型指定返回值为 *online.Player 且返回值个数为1。
	p.Conn.Send([]byte("奖励"))
}

```

最重要的一点，服务启动时要注册创建好的service！
--------------------------------------

`savior.Start(online.Service,campaign.Service)`
```go
package main

import (
    "github.com/mylilcat/savior"
    "github.com/mylilcat/savior/net"
    "online"
    "campaign"
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
    savior.Start(online.Service,campaign.Service) //注册service，service按注册顺序初始化。类比skynet。
}
```


如何使用Service actor调用器 
-----------------------
奈何golang语法限制，service间的通信调用形式想完全仿照skynet，只能通过使用泛型方法尽力贴合了。
没有返回值为Call，根据返回值个数依次为Call1，Call2，Call3，最大支持到Call5。

```go
// Call call other service actor methods with no return value.
// 调用其他服务actor方法，无返回值。
func Call(serviceName string, funcName string, args ...any) {
    targetActor := getServiceActor(serviceName)
    if targetActor == nil {
        return
    }
    task := makeTask(funcName, false, args)
    targetActor.send(task)
}

// Call1 call other service actor methods with 1 return value.
// 调用其他服务actor方法，有一个返回值。
func Call1[R any](serviceName string, funcName string, args ...any) (r R) {
    targetActor := getServiceActor(serviceName)
    if targetActor == nil {
        return
    }
    task := makeTask(funcName, true, args)
    targetActor.send(task)
    result := <-task.resultChan
    r = result[0].(R)
    return
}
```

根据上面的例子看下所需要的参数
* serviceName -> 就是创建service时指定的名称。
* funcName -> 想要调用的actor方法名。
* args ...any actor -> 方法需要的参数。按顺序赋值。

如上述示例  

调用online 无返回值的广播方法   
`service.Call("online","broadcast","活动开始了")`  
"online" -> 服务名  
"broadcast" -> actor方法名  
"活动开始了" -> broadcast需要的参数  

调用online 有一个返回值，获取在线玩家方法  
`p := service.Call1[*online.Player]("online","getPlayer",id)`  
"online" -> 服务名  
"getPlayer" -> actor方法名  
id -> getOnlinePlayerById方法需要的参数    
[*online.Player] -> 返回值要转换的类型。  






