//go:build !windows
// +build !windows

package well

import (
	"bufio"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

// handleSigPipe discards SIGPIPE if the program is running
// as a systemd service.
//
// Background:
//
// systemd interprets programs exited with SIGPIPE as
// "successfully exited" because its default SuccessExitStatus=
// includes SIGPIPE.
// https://www.freedesktop.org/software/systemd/man/systemd.service.html
//
// Normal Go programs ignore SIGPIPE for file descriptors other than
// stdout(1) and stderr(2).  If SIGPIPE is raised from stdout or stderr,
// Go programs exit with a SIGPIPE signal.
// https://golang.org/pkg/os/signal/#hdr-SIGPIPE
//
// journald is a service tightly integrated in systemd.  Go programs
// running as a systemd service will normally connect their stdout and
// stderr pipes to journald.  Unfortunately though, journald often
// dies and restarts due to bugs, and once it restarts, programs
// whose stdout and stderr were connected to journald will receive
// SIGPIPE at their next writes to stdout or stderr.
//
// Combined these specifications and problems all together, Go programs
// running as systemd services often die with SIGPIPE, but systemd will
// not restart them as they "successfully exited" except when the service
// is configured with SuccessExitStatus= without SIGPIPE or Restart=always.
//
// If we catch SIGPIPE and exits abnormally, systemd would restart the
// program.  However, if we call signal.Notify(c, syscall.SIGPIPE),
// SIGPIPE would be raised not only for stdout and stderr but also for
// other file descriptors.  This means that programs that make network
// connections will get a lot of SIGPIPEs and die.  Of course, this is
// not acceptable.
//
// Therefore, we just catch SIGPIPEs and drop them if the program
// runs as a systemd service.  This way, we can detect journald restarts
// by checking the errors from os.Stdout.Write or os.Stderr.Write.
// This check is mainly done in our logger, cybozu-go/log.
func handleSigPipe() {
	if !IsSystemdService() {
		return
	}

	// signal.Ignore does NOT ignore signals; instead, it just stop
	// relaying signals to the channel.  Instead, we set a nop handler.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPIPE)
}

// IsSystemdService returns true if the program runs as a systemd service.
func IsSystemdService() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// https://www.freedesktop.org/software/systemd/man/systemd.exec.html#%24JOURNAL_STREAM
	if len(os.Getenv("JOURNAL_STREAM")) > 0 {
		return true
	}

	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	isService := false
	for sc.Scan() {
		fields := strings.Split(sc.Text(), ":")
		if len(fields) < 3 {
			continue
		}
		if fields[1] != "name=systemd" {
			continue
		}
		isService = strings.HasSuffix(fields[2], ".service")
		break
	}
	if err = sc.Err(); err != nil {
		return false
	}

	return isService
}
