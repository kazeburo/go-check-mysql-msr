package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
	"github.com/mackerelio/checkers"
)

// Version by Makefile
var version string

type opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"5s" description:"Timeout to connect mysql"`
	Crit    int64         `short:"c" long:"critical" description:"critical if seconds behind master is larger than this number"`
	Warn    int64         `short:"w" long:"warning" description:"warning if seconds behind master is larger than this number"`
	Version bool          `short:"v" long:"version" description:"Show version"`
}

type slave struct {
	IORunning   mysqlflags.Bool `mysqlvar:"Slave_IO_Running"`
	SQLRunning  mysqlflags.Bool `mysqlvar:"Slave_SQL_Running"`
	ChannelName *string         `mysqlvar:"Channel_Name"`
	Behind      int64           `mysqlvar:"Seconds_Behind_Master"`
}

func main() {
	ckr := checkMsr()
	ckr.Name = "MySQL Multi Source Replication"
	ckr.Exit()
}

func printVersion() {
	fmt.Printf(`%s %s
Compiler: %s %s
`,
		os.Args[0],
		version,
		runtime.Compiler,
		runtime.Version())
}

func checkMsr() *checkers.Checker {
	opts := opts{}
	psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opts.Version {
		printVersion()
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	db, err := mysqlflags.OpenDB(opts.MyOpts, opts.Timeout, false)
	if err != nil {
		return checkers.Critical(fmt.Sprintf("couldn't connect DB: %v", err))
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	ch := make(chan error, 1)

	var slaves []slave

	go func() {
		e := mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slaves)
		if e != nil {
			ch <- e
			return
		}
		ch <- nil
	}()

	select {
	case err = <-ch:
		// nothing
	case <-ctx.Done():
		err = fmt.Errorf("Connection or query timeout")
	}

	if err != nil {
		return checkers.Critical(fmt.Sprintf("%v", err))
	}

	if len(slaves) == 0 {
		return checkers.Critical("No replication settings")
	}

	var okStatuses []string
	var warnStatuses []string
	var critStatuses []string

	for _, slave := range slaves {
		status := checkers.OK
		if !slave.IORunning.Yes() || !slave.SQLRunning.Yes() {
			status = checkers.CRITICAL
		}
		if opts.Crit > 0 && slave.Behind > opts.Crit {
			status = checkers.CRITICAL
		} else if opts.Warn > 0 && slave.Behind > opts.Warn {
			status = checkers.WARNING
		}

		if slave.ChannelName == nil {
			*slave.ChannelName = "-"
		}

		msg := fmt.Sprintf("%s=io:%s,sql:%s,behind:%d", *slave.ChannelName, slave.IORunning, slave.SQLRunning, slave.Behind)
		switch status {
		case checkers.OK:
			okStatuses = append(okStatuses, msg)
		case checkers.WARNING:
			warnStatuses = append(warnStatuses, msg)
		case checkers.CRITICAL:
			critStatuses = append(critStatuses, msg)
		}
	}

	var msgs []string
	msgs = append(msgs, "[Crit]")
	if len(critStatuses) > 0 {
		msgs = append(msgs, strings.Join(critStatuses[0:], " "))
	} else {
		msgs = append(msgs, "-")
	}
	msgs = append(msgs, "[Warn]")
	if len(warnStatuses) > 0 {
		msgs = append(msgs, strings.Join(warnStatuses[0:], " "))
	} else {
		msgs = append(msgs, "-")
	}
	msgs = append(msgs, "[ OK ]")
	if len(okStatuses) > 0 {
		msgs = append(msgs, strings.Join(okStatuses[0:], " "))
	} else {
		msgs = append(msgs, "-")
	}
	msg := strings.Join(msgs[0:], " ")
	if len(critStatuses) > 0 {
		return checkers.Critical(msg)
	} else if len(warnStatuses) > 0 {
		return checkers.Warning(msg)
	}
	return checkers.Ok(msg)
}
