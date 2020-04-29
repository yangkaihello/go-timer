package timer

import (
	"errors"
	"os/exec"
)

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


type ServerChannelTemplate interface {
	run(host string,token string) ServerChannelTemplate
	Shell(str string) (string,error)
	Close() error
}

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


