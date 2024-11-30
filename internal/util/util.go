package util

import (
	"log"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func FatalError(format string, v ...any) {
	log.Fatalf(format+"\n", v...)
}

type Addr struct {
	Scheme   string
	Hostname string
	Port     string
}

func (a *Addr) Host() string {
	return a.Hostname + ":" + a.Port
}
