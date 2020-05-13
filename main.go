package main

import (
	"golang/timer/timer"
	"log"
)

var testMap = map[string]int{}

func main()  {

	log.Println("Start app path:"+timer.GetAppPath())

	//加载历史任务
	if err := timer.FileValueReader(); err != nil {
		log.Println(err);return
	}

	//定时任务协程守护
	timer.TimeTask()


	//tcp协议加载内容
	tcp := new(timer.TCPTask)
	tcp.Accept()

}
