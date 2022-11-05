package towerslack

import (
	"github.com/tigorlazuardi/tower"
	"github.com/tigorlazuardi/tower-go/towerslack/block"
)

type Templater interface {
	Template(msg tower.MessageContext) block.Blocks
}

type TemplateFunc func(msg tower.MessageContext) block.Blocks

func (f TemplateFunc) Template(msg tower.MessageContext) block.Blocks {
	return f(msg)
}

func defaultTemplate(msg tower.MessageContext) block.Blocks {
	blocks := make(block.Blocks, 0, 20)

	return blocks
}
