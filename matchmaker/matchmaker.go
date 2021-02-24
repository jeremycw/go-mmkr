package matchmaker

import (
	"container/heap"
	"container/list"
	"github.com/google/uuid"
	"time"
)

type MatchConfig struct {
	MinMatchSize   int
	MaxMatchSize   int
	MatchTimeoutMs int
}

type matchMaker struct {
	heap        IntHeap
	scoreToIds  map[int]*list.List
	idToWatcher map[uuid.UUID]watcher
	minSize     int
	matchSize   int
	timeoutMs   int
	time        int
	count       int
}

type watcher struct {
	channel chan uuid.UUID
	matchId uuid.UUID
}

func matchServer(channel chan MatchCmd, conf MatchConfig) {
	mmkr := newMatchMaker(conf.MinMatchSize, conf.MaxMatchSize, conf.MatchTimeoutMs)
	for cmd := range channel {
		cmd.exec(mmkr)
	}
}

func tickMatchMaker(channel chan MatchCmd, tickMs int) {
	for {
		duration := time.Duration(tickMs) * time.Millisecond
		time.Sleep(duration)
		cmd := new(tickCmd)
		*cmd = tickCmd{delta: tickMs}
		channel <- cmd
	}
}

func Start(conf MatchConfig, tickMs int) chan MatchCmd {
	channel := make(chan MatchCmd)
	go matchServer(channel, conf)
	go tickMatchMaker(channel, tickMs)
	return channel
}

func newMatchMaker(minSize int, matchSize int, timeoutMs int) *matchMaker {
	mmkr := new(matchMaker)
	heap.Init(&mmkr.heap)
	mmkr.scoreToIds = make(map[int]*list.List)
	mmkr.idToWatcher = make(map[uuid.UUID]watcher)
	mmkr.matchSize = matchSize
	mmkr.minSize = minSize
	mmkr.timeoutMs = timeoutMs
	mmkr.time = timeoutMs
	return mmkr
}

func (self *matchMaker) join(id uuid.UUID, score int) {
	self.count += 1
	heap.Push(&self.heap, score)
	_, ok := self.scoreToIds[score]
	self.idToWatcher[id] = watcher{}
	if !ok {
		l := list.New()
		l.PushBack(id)
		self.scoreToIds[score] = l
	} else {
		self.scoreToIds[score].PushBack(id)
	}
}

func (self *matchMaker) newMatch() {
	matchId := uuid.New()
	for i := 0; i < self.matchSize && self.heap.Len() > 0; i++ {
		score := heap.Pop(&self.heap).(int)
		// Remove front of list for FIFO ordering
		ids := self.scoreToIds[score]
		idEl := ids.Front()
		ids.Remove(idEl)
		id := idEl.Value.(uuid.UUID)

		w := self.idToWatcher[id]
		if w.matchId == uuid.Nil && w.channel == nil {
			// Empty watcher means the client has joined but not made a /match
			// request yet.
			self.idToWatcher[id] = watcher{matchId: matchId, channel: nil}
		} else if w.matchId == uuid.Nil && w.channel != nil {
			// Client is waiting for match.
			w.channel <- matchId
			delete(self.idToWatcher, id)
		} else {
			panic("Error: Corrupt matchmaker")
		}
		self.count -= 1
	}
}

func (self *matchMaker) watch(id uuid.UUID, channel chan uuid.UUID) {
	w, ok := self.idToWatcher[id]
	if !ok {
		channel <- uuid.Nil
		return
	}
	if w.matchId != uuid.Nil {
		channel <- w.matchId
		delete(self.idToWatcher, id)
	} else {
		self.idToWatcher[id] = watcher{channel: channel}
	}
}

func (self *matchMaker) match() {
	for i := 0; i < self.count/self.matchSize; i++ {
		self.newMatch()
	}
}

func (self *matchMaker) timeout() {
	heap.Init(&self.heap)
	self.count = 0
	self.scoreToIds = make(map[int]*list.List)
	for _, w := range self.idToWatcher {
		if w.channel != nil {
			w.channel <- uuid.Nil
		}
	}
	self.idToWatcher = make(map[uuid.UUID]watcher)
}

func (self *matchMaker) tick(delta int) {
	self.time -= delta
	if self.count >= self.matchSize {
		self.match()
	} else if self.time <= 0 {
		self.time = self.timeoutMs
		if self.count < self.matchSize {
			self.timeout()
		} else {
			self.newMatch()
		}
	}
}
