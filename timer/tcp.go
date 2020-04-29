package timer

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

type TCPTask struct {

}

func (this *TCPTask) Accept() {
	var listener net.Listener
	var err error

	if listener,err = net.Listen("tcp",":8484"); err != nil {
		log.Println(err)
		os.Exit(0)
	}

	//tcp 单线程通信
	for  {
		var taskNumber = 0
		var conn net.Conn
		if conn,err = listener.Accept(); err != nil {
			DebugLog(err.Error(),true);continue
		}
		for _,v := range TimeTaskMap {
			taskNumber += len(v)
		}

		this.addressReader(conn)
	}

}


func (this *TCPTask) addressReader(conn net.Conn) {
	//初始化单线程全局变量
	var err error
	var ReadString string
	var bufReader *bufio.Reader
	//单线程闭包方法
	var Close = func(message string) {
		conn.Write([]byte(strings.Trim(message,"\n")+"\n"))
		conn.Close()
	}
	//读取协议行
	bufReader = bufio.NewReader(conn)
	if ReadString,err = bufReader.ReadString('\n'); err != nil {
		Close(err.Error())
	}

	ReadString = strings.ReplaceAll(ReadString,"\t","")
	ReadString = strings.ReplaceAll(ReadString,"\r","")
	ReadString = strings.ReplaceAll(ReadString,"\n","")

	if ReadString == "log" {
		var logAll = []string{}
		logAll = append(logAll,"go number:" + strconv.Itoa(int(GoTaskNumber)))
		Close(strings.Join(logAll,"\n"));return
	}

	//root@localhost -t=20200424135100 --CMD="ls"
	var timer = new(Timer)

	if err = timer.AnalysisParams(ReadString); err != nil {
		Close(err.Error());return
	}

	//命令行验证
	if d,ok := timer.Params[AGREE_PARAMS_CMD]; !ok || d == "" {
		Close("缺少 --CMD 参数");return
	}

	//定时任务验证,由于是协程等待的定时脚本存在安全性所以需要 时间限制，最大任务数限制
	if date,ok := timer.Params[AGREE_PARAMS_TIMING]; !ok || len(date) != 14 {
		Close("缺少 b  =-t 参数;或是定时格式不正确需要(YmdHis)");return
	}

	TimeTaskMap.Write(timer.Params[AGREE_PARAMS_TIMING],timer)
	atomic.AddInt32(&GoTaskNumber,1)
	Close("已经加入等待队列")

}
