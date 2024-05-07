package shell

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	ginkgo "github.com/onsi/ginkgo/v2"
)

var localenvPath string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	localenvPath = filepath.Join(dir, "../../../localenv")
}

// Run a command in the context of /localenv
func LocalenvCmd(ctx context.Context, cmd string) error {
	// first print the working directory
	return Exec(ctx, fmt.Sprintf("cd %s && %s", localenvPath, cmd))
}

// Run a command with | or other things that do not work in shellcmd
func Exec(ctx context.Context, cmd string) error {
	execCmd := getShellCmd(ctx, cmd)
	return execCmd.Run()
}

func ExecAll(ctx context.Context, cmds ...string) error {
	for _, cmd := range cmds {
		err := Exec(ctx, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func Output(ctx context.Context, cmd string) ([]byte, error) {
	execCmd := getShellCmd(ctx, cmd)

	return execCmd.Output()
}

func PipeInExec(ctx context.Context, cmd string, stdin []byte) error {
	execCmd := getShellCmd(ctx, cmd)
	writer, err := execCmd.StdinPipe()
	if err != nil {
		return err
	}
	writer.Write(stdin)
	writer.Close()

	return execCmd.Run()
}

func PipeInForOutput(ctx context.Context, cmd string, stdin []byte) ([]byte, error) {
	execCmd := getShellCmd(ctx, cmd)
	writer, err := execCmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	writer.Write(stdin)
	writer.Close()

	return execCmd.Output()
}

func getShellCmd(ctx context.Context, cmd string) *exec.Cmd {
	ginkgo.GinkgoWriter.Printf("\033[35m>\033[0m\033[1m %s\033[0m\n", cmd)
	execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	execCmd.Stdout = ginkgo.GinkgoWriter
	execCmd.Stderr = ginkgo.GinkgoWriter
	return execCmd
}
