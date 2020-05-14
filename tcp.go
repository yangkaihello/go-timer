package timer

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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

//tcp 通信协议过程
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

	//额外的字符串操作
	if len(ReadString) > 5 && ReadString[0:6] == "status" {
		go Status(ReadString,conn);return
	} else if len(ReadString) > 5 && ReadString[0:6] == "delete" {
		go Delete(ReadString,conn);return
	} else if len(ReadString) > 5 && ReadString[0:6] == "record" {
		go Record(ReadString,conn);return
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
	Close("已经加入等待队列")

}

const SPLIT_SEARCH = 0x3A 	//参数和数据的区分
const SPLIT_KEYS = 0x2E		//参数的类型区分
const SPLIT_VALUE = 0x2C	//内容的数据区分
const SPLIT_VALUES = 0x7C	//内容中的多个数据区分
const SPLIT_KEY_VALUES_EQ = 0x3D //确认key value 关系的符号

func Status(ReadString string,conn net.Conn) {
	//每个时间段的任务量统计
	var taskDateNumber []string
	for k,v := range TimeTaskMap {
		taskDateNumber = append(taskDateNumber, k+string(SPLIT_KEY_VALUES_EQ)+strconv.Itoa(len(v)))
	}

	var status = map[string]string{
		"task.number": strconv.Itoa(int(GoTaskNumber)),
		"task.date.number": strings.Join(taskDateNumber,string(SPLIT_VALUE)),
	}

	var ReadSlice = strings.Split(ReadString,string(SPLIT_SEARCH))
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
		message += k+string(SPLIT_SEARCH)+v+"\n"
	}

	conn.Write([]byte(message))
	conn.Close()
}

//查看任务的ID
func Record(ReadString string,conn net.Conn) {

	var message []string
	var ids []string
	var ReadSlice = strings.Split(ReadString,string(SPLIT_SEARCH))

	if len(ReadSlice) == 2 { //判断客户端是否想要获取一个状态

		//查看任务队列的数据结构
		if ReadSlice[1] == "all" {
			for date,tasks := range TimeTaskMap {
				for id,_ := range tasks {
					ids = append(ids, strconv.Itoa(id))
				}
				message = append(message, date+string(SPLIT_KEY_VALUES_EQ)+strings.Join(ids,string(SPLIT_VALUES)))
				ids = []string{}
			}
		}else if tasks := TimeTaskMap.Read(ReadSlice[1]); len(tasks) != 0 {
			for id,_ := range tasks {
				ids = append(ids, strconv.Itoa(id))
			}
			message = append(message, ReadSlice[1]+string(SPLIT_KEY_VALUES_EQ)+strings.Join(ids,string(SPLIT_VALUES)))
		}

	}

	//给用户返回删除的数量
	conn.Write([]byte(strings.Join(message,string(SPLIT_VALUE))+"\n"))
	conn.Close()

}

//清除任务队列
func Delete(ReadString string,conn net.Conn) {
	var message string
	var ReadSlice = strings.Split(ReadString,string(SPLIT_SEARCH))
	if len(ReadSlice) == 2 { //判断客户端是否想要获取一个状态
		//对于的数据键删除
		keys := strings.Split(ReadSlice[1],string(SPLIT_KEYS))

		if len(keys) == 1 {
			tasks := TimeTaskMap.ReadDelete(keys[0])
			message = strconv.Itoa(len(tasks))
		} else if len(keys) == 2 {
			//确认第二参数是正确的
			if number,err := strconv.Atoi(keys[1]); err == nil{
				if TimeTaskMap.Delete(keys[0],number) {
					message = "1"
				}else{
					message = "0"
				}
			}
		}

	}

	//给用户返回删除的数量
	conn.Write([]byte(message+"\n"))
	conn.Close()
}
