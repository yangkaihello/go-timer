package timer

import (
	"errors"
	"strings"
)


type Timer struct {
	Host string
	Token string
	Params map[string]string
	Str string
}

func (this *Timer) AnalysisParams(str string) error {
	var linuxShell []string
	var linuxShellHost []string

	this.Params = map[string]string{
		AGREE_PARAMS_TIMING:"",
		AGREE_PARAMS_CMD:"",
	}

	this.Str = str
	linuxShell = strings.Fields(str)

	if len(linuxShell) < 1 || len(strings.Split(linuxShell[0],"@")) != 2 {
		return errors.New("数据结构错误")
	}

	linuxShellHost = strings.Split(linuxShell[0],"@")

	this.Host = linuxShellHost[1]
	this.Token = linuxShellHost[0]

	linuxShell = linuxShell[1:]
	linuxShell = ParamsSegmentAnalysis(linuxShell)

	for _,v := range linuxShell{
		params := strings.Split(v,"=")
		if len(params) != 2{
			continue
		}

		if _,ok := this.Params[params[0]]; ok {
			this.Params[params[0]] = strings.Trim(params[1],"\"")
		}
	}

	return nil
}

func (this *Timer) GetHost() (string,string) {
	return this.Host,this.Token
}

func (this *Timer) GetParam(str string) string {
	return this.Params[str]
}
