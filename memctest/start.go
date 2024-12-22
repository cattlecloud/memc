// Copyright (c) CattleCloud LLC
// SPDX-License-Identifier: BSD-3-Clause

package memctest

import (
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/test/must"
	"github.com/shoenig/test/portal"
	"github.com/shoenig/test/skip"
	"github.com/shoenig/test/wait"
)

const (
	executable = "memcached"
)

type tester struct{}

func (ft *tester) Fatalf(msg string, args ...any) {
	s := fmt.Sprintf(msg, args...)
	panic(s)
}

var (
	fatal = new(tester)
	ports = portal.New(fatal)
)

func LaunchTCP(t *testing.T, args []string) (string, func()) {
	// requires memcached executable on $PATH
	skip.CommandUnavailable(t, executable)

	port := ports.One()
	address := fmt.Sprintf("localhost:%d", port)
	args = append(args, "-l", address)

	ctx, cancel := scope.Cancelable()
	cmd := exec.CommandContext(ctx, executable, args...)
	err := cmd.Start()
	must.NoError(t, err)

	// wait for memcached to be listening
	must.Wait(t, wait.InitialSuccess(
		wait.Timeout(3*time.Second),
		wait.Gap(200*time.Millisecond),
		wait.ErrorFunc(func() error {
			_, err := net.Dial("tcp", address)
			return err
		}),
	))

	return address, cancel
}
