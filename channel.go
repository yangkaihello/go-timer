package timer

import (
	"errors"
	"os/exec"
)

//服务器连接通道
type Channel struct {
	timer *Timer
	Server ServerChannelTemplate
}

func (this *Channel) Run(timer *Timer) error {
	this.timer = timer

	if this.timer.Host == "localhost" || this.timer.Host == "127.0.0.1" {
		this.Server = new(ServerLocalhost).run(this.timer.Host,this.timer.Token)
	}else{
		return errors.New("not agent server")
	}
	return nil
}

//服务器通信类型接口
type ServerChannelTemplate interface {
	//初始化数据，提供host,token
	//返回本身接口
	run(host string,token string) ServerChannelTemplate
	//代理执行 shell 数据
	Shell(str string) (string,error)
	//关闭数据的连接
	Close() error
}

//本地的服务器代理连接
//必须满足 ServerChannelTemplate 接口
type ServerLocalhost struct {
	client *exec.Cmd
	host string
	token string
}

func (this *ServerLocalhost) run(host string,token string) ServerChannelTemplate {
	this.host = host
	this.token = token

	return this
}

func (this *ServerLocalhost) Shell(str string) (string,error) {
	this.client = exec.Command("/bin/bash","-c",str)
	bytes,err := this.client.CombinedOutput()
	return string(bytes),err
}

func (this *ServerLocalhost) Close() error {
	return nil
}


