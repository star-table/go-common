package goroutine

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/star-table/go-common/go-common/utils/stack"
)

func SafeRun(fn func(), logger *log.Helper) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("err:%v, stack:%v", r, stack.GetStack())
			}
		}()
		fn()
	}()
}
