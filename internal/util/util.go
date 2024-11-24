package util

import "log"

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func FatalError(format string, v ...any) {
	log.Fatalf(format+"\n", v...)
}
