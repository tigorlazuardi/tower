package bucket

import (
	"github.com/bwmarrin/snowflake"
	"os"
)

var snowflakeNode *snowflake.Node

func init() {
	host, _ := os.Hostname()
	snowflakeNode = generateSnowflakeNodeFromString("tower-bucket-" + host)
}

type LengthHint interface {
	Len() int
}

func generateSnowflakeNodeFromString(s string) *snowflake.Node {
	id := 0
	for _, c := range s {
		id += int(c)
	}
	for id > 1023 {
		id >>= 1
	}
	node, _ := snowflake.NewNode(int64(id))
	return node
}
