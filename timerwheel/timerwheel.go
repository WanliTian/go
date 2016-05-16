package timerwheel

import (
	"errors"
	"sync"
	"time"
)

var ErrTooLarge error = errors.New("duration too large for timer wheel")

type TimerWheel struct {
	lock   sync.Mutex
	hour   int64
	minute int64
	second int64
	events map[int64]map[int64]map[int64]chan struct{}
	close  chan struct{}
}

func NewTimerWheel() *TimerWheel {
	return &TimerWheel{
		events: make(map[int64]map[int64]map[int64]chan struct{}),
	}
}

func (this *TimerWheel) Serve() {
	t := time.NewTimer(time.Second)
	for {
		select {
		case <-t.C:
			this.lock.Lock()
			this.second = (this.second + 1) % 60
			if this.second == 0 {
				this.minute = (this.minute + 1) % 60
				if this.minute == 0 {
					this.hour = (this.hour + 1) % 24
				}
			}
			channel, ok := this.events[this.hour][this.minute][this.second]
			if ok {
				close(channel)
				delete(this.events[this.hour][this.minute], this.second)
			}
			this.lock.Unlock()
			t.Reset(time.Second)
		case <-this.close:
			t.Stop()
			close(this.close)
			return
		}
	}
}

func (this *TimerWheel) After(dur time.Duration) chan struct{} {
	if dur >= 24*time.Hour {
		panic(ErrTooLarge)
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	h := (int64(dur.Seconds()/60)/60 + this.hour) % 24
	m := ((int64(dur.Seconds())/60)%60 + this.minute) % 60
	s := ((int64(dur.Seconds()) % 60) + this.second) % 60

	mMapper, ok := this.events[h]
	if !ok {
		mMapper = make(map[int64]map[int64]chan struct{})
		this.events[h] = mMapper
	}
	sMapper, ok := mMapper[m]
	if !ok {
		sMapper = make(map[int64]chan struct{})
		mMapper[m] = sMapper
	}
	channel, ok := sMapper[s]
	if !ok {
		channel = make(chan struct{})
		sMapper[s] = channel
	}
	return channel
}

func (this *TimerWheel) Stop() {
	this.close <- struct{}{}
	<-this.close
}
