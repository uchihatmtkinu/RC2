package main

import (
	"testing"
	"time"

	"github.com/uchihatmtkinu/RC/testforclient/network"
)

func TestServer(t *testing.T) {
	network.StartLocalServer()
}

func TestTimer(t *testing.T) {
	now := time.Now()
	next := now.Add(0)
	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), next.Minute()+1, 0, 0, next.Location())
	xx := next.Sub(now)
	tt := time.NewTimer(next.Sub(now))
	<-tt.C
	t.Error(time.Now())
	t.Error(xx)

}
