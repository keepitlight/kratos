package runtime

import (
	"context"
)

// Routine 表示可执行对象
type Routine func(ctx context.Context) error
