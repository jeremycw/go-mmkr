package main

import (
	"github.com/google/uuid"
)

type matchCmd interface {
	exec(executor matchCmdExecutor)
}

type matchCmdExecutor interface {
	execJoin(cmd joinCmd)
	execTick(cmd tickCmd)
	execWatch(cmd watchCmd)
}

type joinCmd struct {
	id    uuid.UUID
	score int
}

type watchCmd struct {
	id      uuid.UUID
	channel chan uuid.UUID
}

type tickCmd struct {
	delta int
}

func (self *matchMaker) execJoin(cmd joinCmd) {
	self.join(cmd.id, cmd.score)
}

func (self *matchMaker) execWatch(cmd watchCmd) {
	self.watch(cmd.id, cmd.channel)
}

func (self *matchMaker) execTick(cmd tickCmd) {
	self.tick(cmd.delta)
}

func (self joinCmd) exec(executor matchCmdExecutor) {
	executor.execJoin(self)
}

func (self watchCmd) exec(executor matchCmdExecutor) {
	executor.execWatch(self)
}

func (self tickCmd) exec(executor matchCmdExecutor) {
	executor.execTick(self)
}
