package util

import (
	"fmt"
	"os"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func FatalError(format string, v ...any) {
	fmt.Printf("socks2http: "+format+"\n", v...)
	os.Exit(1)
}
