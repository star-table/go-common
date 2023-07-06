package errors_test

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	//var t *testing.T
	fmt.Println("TestMain")

	//todo:后期优化成httptesting.NewServer 模式
	//sm := http.NewServeMux()
	//httpcli = httptesting.NewServer(g, false)
	//s := httptest.NewServer(g)
	//defer s.Close()

	code := m.Run()

	//httpcli.Close()

	os.Exit(code)
}
