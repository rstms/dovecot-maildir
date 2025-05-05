package cmd

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os/exec"
	"testing"
)

func run(t *testing.T, command string, args ...string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	require.Nil(t, err)
	fmt.Println(string(out))
	require.Nil(t, err)
}

func TestInit(t *testing.T) {
	run(t, "rm", "-rf", "testdata/Maildir")
	run(t, "cp", "-rp", "testdata/src", "testdata/Maildir")
}
