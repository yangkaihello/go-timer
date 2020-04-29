# go-timer
go 的任务调度器可以代理执行本地服务器的脚本 如下是一个任务开启的方式 

```
//加载历史任务
if err := timer.FileValueReader(); err != nil {
    log.Println(err);return
}

//定时任务协程守护
timer.TimeTask()

//tcp协议加载内容
tcp := new(timer.TCPTask)
tcp.Accept()
```

### 通信任务格式
```
//未来目标是准备设定一个任务调取管理器，会有主程序以及子程序
//通过主程序的服务器 来选择性的调度到对于子程序的服务器
"token@localhost -t=20200429121700 --CMD=\"ls\""
//服务器
token@localhost

//时间参数 一个纯数字的14位时间YmdHis
-t=20200429121700

//代理执行命令行 需要用双引号来代表开始和结束 内置的字符串需要用单引号来操作
--CMD="ls | grep 't'"

```

### 客户端使用TCP通信来创建定时任务 golang 实例（请使用各语言的tcp通信实例）

```
//申明一个io.reader 读取器
var reader *bufio.Reader
//连通tcp 8484 的本地端口
conn,_ := net.Dial("tcp","127.0.0.1:8484")
//创建一个io.reader读取器
reader = bufio.NewReader(conn)
//写入通信内容
bs := []byte(`root@localhost -t=20200429121700 --CMD="ls | grep 't'"`+"\n")
conn.Write(bs)
//读取通信内容
log.Print(ByteReader.ReadString('\n'))
//关闭连接
conn.Close()
```

