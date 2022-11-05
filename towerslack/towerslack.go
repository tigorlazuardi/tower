package towerslack

import "github.com/tigorlazuardi/tower"

type Slack struct {
	token   string
	channel string
	tracer  tower.TraceCapturer
}
