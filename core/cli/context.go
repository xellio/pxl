package cli

import (
	"fmt"
	"time"
)

type Context struct {
	IsEncodeMode bool
	IsDecodeMode bool
	IsDebugMode  bool
	Source       string
	Target       string
}

func (c Context) DebugString() string {
	return fmt.Sprintf(`%s
isDebugMode	: %t
isEncode 	: %t
isDecode 	: %t
source   	: %s
target   	: %s`,
		time.Now().Local(),
		c.IsEncodeMode,
		c.IsDecodeMode,
		c.IsDebugMode,
		c.Source,
		c.Target,
	)
}
