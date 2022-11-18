package towerhttp

import (
	"github.com/tigorlazuardi/tower"
)

type TowerHttp struct {
	encoder    Encoder
	transform  BodyTransform
	tower      *tower.Tower
	compressor Compression
}
