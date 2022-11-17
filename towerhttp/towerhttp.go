package towerhttp

import (
	"context"
	"github.com/tigorlazuardi/tower"
	"net/http"
)

type TowerHttp struct {
	encoder   Encoder
	transform BodyTransform
	tower     *tower.Tower
}

func (t TowerHttp) Respond(ctx context.Context, rw http.ResponseWriter, body any) {
}
