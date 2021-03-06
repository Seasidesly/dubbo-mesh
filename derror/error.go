// 错误处理相关

package derror

import (
	"dubbo-mesh/log"
)

func Warn(err error) {
	if err != nil {
		log.Warn(err.Error())
	}
}

func Panic(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}
