package timer

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const AGREE_PARAMS_TIMING = "-t"
const AGREE_PARAMS_CMD = "--CMD"


type timeTaskMapTemplate map[string][]*Timer

var TimeChan = make(chan *Timer)	//协程任务调度器,常驻内存重启丢失
var GoTaskNumber int32 = 0	//协程任务,计数器

//轮询定时器
var TimeTaskMap = make(timeTaskMapTemplate)	//时间任务调度器，零时存放在内存中，通过 TimeTaskMapFileValueStore | TimeTaskMapFileValueTime 来控制持久储存
var timeTaskMapFileValue = make(map[string]string) //持久化储存数据结构


//持久化储存模式
var timeTaskDBFile = GetAppPath()+"/.task.db"
var timeTaskMapFileValueNumber int32 = 0 //持久化储存数据，计数器
var timeTaskMapTimeoutExec time.Duration = 30 //超时任务重新执行的轮询间隔时间

var TimeTaskMapFileValueMode = 0 //持久化储存模式 1：任务计数器模式，2：时间模式，默认时间模式
var TimeTaskMapFileValueTime = 60 //持久化储存数据，触发器，多少秒执行一次
var TimeTaskMapFileValueStore int32 = 100 //持久化储存数据，触发器 100个任务储存一次

//log 记录配置
var LogMode = 2	//log记录模式 1：显示器输出，2：文件记录模式，无记录
var LogFilePath = GetAppPath()+"/operation.log"		//异常记录日志
var LogMustFilePath = GetAppPath()+"/important.log"	//必须记录日志的文件
var LogFileSuccess = ""					//成功日志

type LogChanTemplate struct {
	FileName string
	Content string
}
var LogChan = make(chan LogChanTemplate)
var Lock sync.Mutex

func (this *timeTaskMapTemplate) Write(key string,value *Timer) {
	Lock.Lock()
	(*this)[key] = append((*this)[key], value)
	timeTaskMapFileValue[key] += value.Str+"\n"
	timeTaskMapFileValueNumber++
	Lock.Unlock()
}

func (this *timeTaskMapTemplate) Read(key string) []*Timer {
	Lock.Lock()
	var SliceStr []*Timer
	var ok bool
	if SliceStr,ok = (*this)[key];ok {
		delete(timeTaskMapFileValue,key)
		delete(*this,key)
		timeTaskMapFileValueNumber = timeTaskMapFileValueNumber+int32(len(SliceStr))
	}
	Lock.Unlock()
	return SliceStr
}

func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))
	return path[:index]
}

//文件储存自检器
func fileValue()  {
	var buf = new(bytes.Buffer)
	for _,v := range timeTaskMapFileValue {
		buf.WriteString(v)
	}
	ioutil.WriteFile(timeTaskDBFile,buf.Bytes(),os.ModePerm)
}

//任务加载器
func FileValueReader() error {
	var err error
	var bytes []byte

	//自动创建文件
	if _,err := os.Stat(timeTaskDBFile); err != nil {
		_,_ = os.Create(timeTaskDBFile)
	}

	bytes,err = ioutil.ReadFile(timeTaskDBFile)
	data := strings.Split(string(bytes),"\n")

	for _,v := range data {
		if v == "" {
			continue
		}
		var timer = new(Timer)
		err = timer.AnalysisParams(v)
		TimeTaskMap.Write(timer.GetParam(AGREE_PARAMS_TIMING),timer)
		atomic.AddInt32(&GoTaskNumber,1)
	}
	return err
}

//文件的创建并连续写入
func WriteFile()  {

	for template := range LogChan {
		var file *os.File
		if _,err := os.Stat(template.FileName); err != nil {
			file,_ = os.Create(template.FileName)
		}else{
			file,_ = os.OpenFile(template.FileName,os.O_WRONLY,0666)
		}
		n,_ := file.Seek(0,os.SEEK_END)
		file.WriteAt([]byte(template.Content),n)
		file.Close()
	}

}

//str string 日志内容无需换行符号
//isMust bool 是否是重要信息重要信息记录到文件日志中
//全局变量 logMode 控制输出模式 1 = io输出,2 = os文件输出 其余类型表示关闭日志
func DebugLog(str string,isMust bool) {

	str = strings.Trim(str,"\n")
	//必须记录的日志采用文件日志
	if isMust == true {
		LogChan <- LogChanTemplate{FileName:LogMustFilePath,Content:"[important]"+str+"; "+time.Now().Format("2006-01-02 15:04:05")+"\n"}
	}else{
		switch LogMode {
		case 1:
			log.Println(str)
		case 2:
			LogChan <- LogChanTemplate{FileName:LogFilePath,Content:time.Now().Format("2006-01-02 15:04:05")+":"+str+"\n"}
		}
	}

}

//YmdHis 格式日期 拆分
func NumberDateAnalysis(date string) (
	year int,
	month int,
	day int,
	hour int,
	min int,
	sec int,
) {
	if len(date) != 14 {
		return
	}
	year, _ = strconv.Atoi(date[0:4])
	month,_ = strconv.Atoi(date[4:6])
	day,_ = strconv.Atoi(date[6:8])
	hour,_ = strconv.Atoi(date[8:10])
	min,_ = strconv.Atoi(date[10:12])
	sec,_ = strconv.Atoi(date[12:14])
	return
}

//获取 YmdHis 格式日期
func NumberDate() ( datetime string ) {
	return time.Now().Format("20060102150405")
}

//解析cmd 参数
func ParamsSegmentAnalysis(SliceStr []string) []string {

	var isContinuous = false
	var linuxShell []string

	for _,v := range SliceStr{
		if isContinuous == true && !strings.ContainsAny(v,"\"") {
			linuxShell[len(linuxShell)-1] = linuxShell[len(linuxShell)-1]+" "+v
			continue
		} else if isContinuous == true && strings.ContainsAny(v,"\"") {
			linuxShell[len(linuxShell)-1] = linuxShell[len(linuxShell)-1]+" "+v
			isContinuous = false
			continue
		}

		if strings.ContainsAny(v,"\"") && strings.Count(v,"\"") != 2 {
			isContinuous = true
			linuxShell = append(linuxShell, v)
		}else{
			linuxShell = append(linuxShell, v)
		}

	}
	return linuxShell

}

//定时任务需要的go 协程调度
func TimeTask() {

	//日志等待写入
	go func() {
		WriteFile()
	}()

	//文件储存协程自检器
	go func() {
		for {
			switch TimeTaskMapFileValueMode {
			case 1:
				if timeTaskMapFileValueNumber >= TimeTaskMapFileValueStore ||
					len(timeTaskMapFileValue) == 0 {
					fileValue()
					timeTaskMapFileValueNumber=0
				}
			case 2:
				if time.Now().Unix()%int64(TimeTaskMapFileValueTime) == int64(TimeTaskMapFileValueTime-1) {
					fileValue()
					time.Sleep(time.Second*1)
				}
			default:
				if time.Now().Unix()%int64(TimeTaskMapFileValueTime) == int64(TimeTaskMapFileValueTime-1) {
					fileValue()
					time.Sleep(time.Second*1)
				}
			}

			time.Sleep(time.Millisecond*100)
		}

	}()

	//超时任务调度器 每隔60秒重新检测是否存在未调度的任务
	go func() {
		for  {
			var date = NumberDate()
			for k,_ := range TimeTaskMap {
				if k < date {
					for _,task := range TimeTaskMap.Read(k) {
						TimeChan <- task
					}
				}
			}
			time.Sleep(time.Second*timeTaskMapTimeoutExec)
		}
	}()

	//时间任务调度器
	//每0.5 秒调度一下协程查看是否需要开辟任务
	go func() {

		for  {
			datetime := NumberDate()
			data := TimeTaskMap.Read(datetime)
			if data != nil{
				for _,Record := range data {
					go func() {
						var data string
						var err error

						//任务执行地
						var channel = new(Channel)
						if err := channel.Run(Record); err != nil {
							DebugLog(Record.Str+":"+err.Error(),false)
							TimeTaskMap.Write(Record.GetParam(AGREE_PARAMS_TIMING),Record)
							return
						}

						if data,err = channel.Server.Shell(Record.GetParam(AGREE_PARAMS_CMD)); err != nil {
							DebugLog(Record.Str+":"+err.Error(),false)
							TimeTaskMap.Write(Record.GetParam(AGREE_PARAMS_TIMING),Record)
							return
						}

						data = strings.ReplaceAll(data,"\r","")
						data = strings.ReplaceAll(data,"\t","")
						data = strings.ReplaceAll(data,"\n","")
						if LogFileSuccess != "" {
							LogChan <- LogChanTemplate{FileName:LogFileSuccess,Content:data+"\n"}
						}
						channel.Server.Close()
						atomic.AddInt32(&GoTaskNumber,-1)
					}()
				}
			}
			time.Sleep(time.Millisecond*500)
		}

	}()

	//超时任务重新管道通信，立即执行 ,不会重新记录防止无限执行
	go func() {

		for Record := range TimeChan{
			var data string
			var err error

			//任务执行地
			var channel = new(Channel)
			if err := channel.Run(Record); err != nil {
				DebugLog("[run]Shell: "+Record.Str+";Error: "+err.Error(),true)
				continue
			}

			if data,err = channel.Server.Shell(Record.GetParam(AGREE_PARAMS_CMD)); err != nil {
				DebugLog("[shell]Shell: "+Record.Str+";Error: "+err.Error(),true)
				continue
			}

			data = strings.ReplaceAll(data,"\r","")
			data = strings.ReplaceAll(data,"\t","")
			data = strings.ReplaceAll(data,"\n","")
			if LogFileSuccess != "" {
				LogChan <- LogChanTemplate{FileName:LogFileSuccess,Content:data+"\n"}
			}
			channel.Server.Close()
			atomic.AddInt32(&GoTaskNumber,-1)

		}
	}()


}
