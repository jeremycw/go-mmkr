package matchmaker

import (
	"github.com/google/uuid"
)

type MatchCmd interface {
	exec(executor matchCmdExecutor)
}

type matchCmdExecutor interface {
	execJoin(cmd JoinCmd)
	execTick(cmd tickCmd)
	execWatch(cmd WatchCmd)
}

type JoinCmd struct {
	Id    uuid.UUID
	Score int
}

type WatchCmd struct {
	Id      uuid.UUID
	Channel chan uuid.UUID
}

type tickCmd struct {
	delta int
}

func (self *matchMaker) execJoin(cmd JoinCmd) {
	self.join(cmd.Id, cmd.Score)
}

func (self *matchMaker) execWatch(cmd WatchCmd) {
	self.watch(cmd.Id, cmd.Channel)
}

func (self *matchMaker) execTick(cmd tickCmd) {
	self.tick(cmd.delta)
}

func (self JoinCmd) exec(executor matchCmdExecutor) {
	executor.execJoin(self)
}

func (self WatchCmd) exec(executor matchCmdExecutor) {
	executor.execWatch(self)
}

func (self tickCmd) exec(executor matchCmdExecutor) {
	executor.execTick(self)
}
