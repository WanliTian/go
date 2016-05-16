package timerwheel

import (
	"errors"
	"sync"
	"time"
)

const (
	ErrNotExist = errors.new("no event finded in timer wheel")
	ErrTooLarge = errors.new("duration too large for timer wheel")
)

type TimerWheel struct {
	lock   sync.Mutex
	hour   int8
	minute int8
	second int8
	events map[int8]map[int8]map[int8]chan struct{}
	close  chan struct{}
}

func NewTimerWheel() *TimerWheel {
	return &TimerWheel{
		events: make(map[int8]map[int8]map[int8]chan struct{}),
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
			close(this.events[this.hour][this.minute][this.second])
			delete(this.events[this.hour][this.minute], this.second)
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
		panic(ErTooLarge)
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	h := ((dur/60)/60 + this.hour) % 24
	m := ((dur/60)%60 + this.minute) % 60
	s := ((dur % 60) + this.second) % 60

	mMapper, ok := this.events[h]
	if !ok {
		mMapper = make(map[int8]map[int8][]*Event)
		this.events[h] = mMapper
	}
	sMapper, ok := mMapper[m]
	if !ok {
		sMapper = make(map[int8][]*Event)
		mMapper[m] = sMapper
	}
	channel, ok := sMapper[s]
	if !ok {
		channel = make(chan struct{})
	}
	return channel
}

func (this *TimerWheel) Stop() {
	this.close <- struct{}{}
	<-this.close
}
