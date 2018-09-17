package services

import (
	"github.com/robfig/cron"
	"git.profzone.net/profzone/peer-to-world/global"
)

var applyTask *cron.Cron

func init() {
	applyTask = cron.New()
	//applyTask.AddFunc(global.Config.HeartbeatTaskSpec, CheckHeartbeatAction)
	applyTask.AddFunc(global.Config.RequestHeightTaskSpec, RequestHeightTask)
}

func StartTask() {
	if applyTask != nil {
		applyTask.Start()
	}
}
