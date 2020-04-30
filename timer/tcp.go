package timer

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
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

		var conn net.Conn
		if conn,err = listener.Accept(); err != nil {
			DebugLog(err.Error(),true);continue
		}

		//超时设置0.01秒就自动关闭连接,用作保护单线程的通信
		var timeNow = time.Now()
		go func() {
			for  {
				if time.Now().Sub(timeNow).Milliseconds() > 10 {
					break
				}
				time.Sleep(time.Millisecond*1)
			}
			conn.Close()
		}()

		//暂时不开放对外通信
		var RemoteAddr = conn.RemoteAddr().String()
		if RemoteAddr[:5] != "[::1]" && strings.Split(RemoteAddr,":")[0] != "127.0.0.1" {
			conn.Close();continue
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
	if ReadString,err = bufReader.ReadString('\n'); err != nil || len(ReadString) == 0 {
		Close(err.Error());return
	}

	ReadString = strings.ReplaceAll(ReadString,"\t","")
	ReadString = strings.ReplaceAll(ReadString,"\r","")
	ReadString = strings.ReplaceAll(ReadString,"\n","")

	if len(ReadString) > 5 && ReadString[0:6] == "status" {
		go Status(ReadString,conn);return
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

func Status(ReadString string,conn net.Conn) {
	//每个时间段的任务量统计
	var taskDateNumber string
	for k,v := range TimeTaskMap {
		taskDateNumber += k+"="+strconv.Itoa(len(v))+","
	}

	var status = map[string]string{
		"task.number": strconv.Itoa(int(GoTaskNumber)),
		"task.date.number": strings.Trim(taskDateNumber,","),
	}
	var ReadSlice = strings.Split(ReadString,":")

	if len(ReadSlice) == 2 { //判断客户端是否想要获取一个状态

		if d,ok := status[ReadSlice[1]]; ok { //判断是否是一个允许访问的状态
			status = map[string]string{
				ReadSlice[1]: d,
			}
		}
	}

	var message string
	//最终结构
	for k,v := range status {
		message += k+":"+v+"\n"
	}

	conn.Write([]byte(message))
	conn.Close()
}
