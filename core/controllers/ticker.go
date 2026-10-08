package controllers

import (
	"github.com/hunterhug/fafacms/core/model"
	log "github.com/hunterhug/golog"
	"time"
)

var LoopChan = make(chan CountType, 1000)

type CountType struct {
	UserId int64
	NodeId int64
	T      int // 1,2,3
}

func SendToLoop(userId, nodeId int64, t int) {
	LoopChan <- CountType{
		UserId: userId,
		NodeId: nodeId,
		T:      t,
	}
}

func LoopCount() {
	log.Debugf("Ticker start")
	for {
		select {
		case v := <-LoopChan:
			log.Debugf("Ticker Count: %#v", v)
			if v.T == 1 {
				if v.UserId != 0 {
					err := model.CountContentAll(v.UserId)
					if err != nil {
						log.Errorf("Ticker Count all content err: %s", err.Error())
					}
				}
			} else if v.T == 2 {
				if v.UserId != 0 && v.NodeId != 0 {
					err := model.CountContentOneNode(v.UserId, v.NodeId)
					if err != nil {
						log.Errorf("Ticker Count node content err: %s", err.Error())
					}
				}
			} else if v.T == 3 {
				if v.UserId != 0 {
					err := model.CountContentCool(v.UserId)
					if err != nil {
						log.Errorf("Ticker Count all content cool err: %s", err.Error())
					}
				}
			}

		case <-time.After(5 * time.Second):
			//log.Debugf("Ticker...")

		}
	}
}
