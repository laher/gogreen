package pass

import (
	"testing"
	"time"
)

func TestSlow(t *testing.T) {
	time.Sleep(time.Second * 1)
}

func TestSlow2(t *testing.T) {
	time.Sleep(time.Second * 2)
}

func TestSlow3(t *testing.T) {
	time.Sleep(time.Second * 3)
}
