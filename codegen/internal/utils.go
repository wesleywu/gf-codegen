package internal

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"os/exec"
)

func ExecCommand(ctx context.Context, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	bytes, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(bytes) > 0 {
		g.Log().Info(ctx, string(bytes))
	} else {
		g.Log().Info(ctx, "done")
	}
	return nil
}

func ImportModule(ctx context.Context, module string) error {
	g.Log().Infof(ctx, "importing %s", module)
	cmd := exec.Command("go", "get", module)
	bytes, err := cmd.Output()
	if err != nil {
		return err
	}
	if len(bytes) > 0 {
		g.Log().Info(ctx, string(bytes))
	} else {
		g.Log().Info(ctx, "done")
	}
	return nil
}
