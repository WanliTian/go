package timerwheel

import (
	"testing"
	"time"
)

func Test_TimerWheelSeconds(t *testing.T) {
	tw := NewTimerWheel()
	go tw.Serve()

	begin := time.Now().Second()
	channel := tw.After(time.Second * 6)
	<-channel
	if (time.Now().Second()-begin+60)%60 != 6 {
		t.Errorf("test timerwheel secons error")
	}
}

func Test_Func2(t *testing.T) {
	_ = time.Second
	t.Log("hello world")
}
