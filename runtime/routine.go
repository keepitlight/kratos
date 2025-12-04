package runtime

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// Routine 表示可执行对象
type Routine func(ctx context.Context, logger log.Logger) error
