package runtime_test

import (
	"context"
	"fmt"
	"time"

	"github.com/keepitlight/kratos/runtime"
)

func ExampleCo() {
	runtime.Co(func(ctx context.Context) error {
		return nil
	})
}

func ExamplePreload() {
	runtime.Preload(func() error {
		return nil
	})
}

func ExampleStart() {
	ctx := context.Background()
	_, err, _ := runtime.Start(ctx, nil, nil, "build", "commit", time.Now())
	if err != nil {
		return
	}
	a, r, b, c, _ := runtime.State()
	fmt.Println(a, r, b, c)
	// output:
	// <nil> <nil> build commit
}
