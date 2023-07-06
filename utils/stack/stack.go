package stack

import "runtime"

func GetStack() string {
	bts := make([]byte, 8192)
	runtime.Stack(bts, false)
	return string(bts)
}
