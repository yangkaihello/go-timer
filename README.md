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

