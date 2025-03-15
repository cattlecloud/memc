// Copyright (c) CattleCloud LLC
// SPDX-License-Identifier: BSD-3-Clause

package memctest

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
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

func waitUntilReady(t *testing.T, mode, address string) {
	must.Wait(t, wait.InitialSuccess(
		wait.Timeout(3*time.Second),
		wait.Gap(200*time.Millisecond),
		wait.ErrorFunc(func() error {
			_, err := net.Dial(mode, address)
			return err
		}),
	))
}

func LaunchTCP(t *testing.T, args []string) (string, func()) {
	// requires memcached executable on $PATH
	skip.CommandUnavailable(t, executable)

	// configure a loopback address to listen on
	port := ports.One()
	address := fmt.Sprintf("localhost:%d", port)
	args = append(args, "-l", address)

	// start the memcached process
	ctx, cancel := scope.Cancelable()
	cmd := exec.CommandContext(ctx, executable, args...)
	err := cmd.Start()
	must.NoError(t, err)

	// wait for memcached to be listening
	waitUntilReady(t, "tcp", address)

	// good to go!
	return address, cancel
}

func LaunchUDS(t *testing.T, args []string) (string, func()) {
	// requires memcached executable on $PATH
	skip.CommandUnavailable(t, executable)

	// configure a socket file to listen on
	dir := t.TempDir()
	socket := filepath.Join(dir, "test.sock")
	args = append(args, "--unix-socket", socket)

	// start the memcached instance
	ctx, cancel := scope.Cancelable()
	cmd := exec.CommandContext(ctx, executable, args...)
	err := cmd.Start()
	must.NoError(t, err, must.Func(func() string {
		b, _ := cmd.CombinedOutput()
		return "unable to start!\n" + string(b)
	}))

	// wait for memcached to be listening
	waitUntilReady(t, "unix", socket)

	// good to go!
	return socket, cancel
}
