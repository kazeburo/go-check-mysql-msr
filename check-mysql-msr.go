package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jessevdk/go-flags"
	"github.com/kazeburo/go-mysqlflags"
	"github.com/mackerelio/checkers"
)

// Version by Makefile
var version string

type Opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"5s" description:"Timeout to connect mysql"`
	Crit    int64         `short:"c" long:"critical" description:"critical if seconds behind master is larger than this number"`
	Warn    int64         `short:"w" long:"warning" description:"warning if seconds behind master is larger than this number"`
	Version bool          `short:"v" long:"version" description:"Show version"`
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
	opts := Opts{}
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

	var okStatuses []string
	var warnStatuses []string
	var critStatuses []string

	go func() {
		rows, e := db.Query("SHOW SLAVE STATUS")
		if e != nil {
			ch <- e
			return
		}
		defer rows.Close()

		cols, e := rows.Columns()
		if e != nil {
			ch <- e
			return
		}
		vals := make([]interface{}, len(cols))
		idxSlaveIORunning := -1
		idxSlaveSQLRunning := -1
		idxChannelName := -1
		idxSecondsBehindMaster := -1
		for i, v := range cols {
			vals[i] = new(sql.RawBytes)
			if v == "Slave_IO_Running" {
				idxSlaveIORunning = i
			}
			if v == "Slave_SQL_Running" {
				idxSlaveSQLRunning = i
			}
			if v == "Channel_Name" {
				idxChannelName = i
			}
			if v == "Seconds_Behind_Master" {
				idxSecondsBehindMaster = i
			}
		}
		if idxSlaveIORunning < 0 || idxSlaveSQLRunning < 0 || idxSecondsBehindMaster < 0 {
			ch <- fmt.Errorf("Could not find Slave_IO_Running or Slave_SQL_Running or Seconds_Behind_Master in columns")
			return
		}

		i := 0
		for rows.Next() {
			i++
			e = rows.Scan(vals...)
			if e != nil {
				ch <- e
				return
			}
			slaveIORunning := string(*vals[idxSlaveIORunning].(*sql.RawBytes))
			slaveSQLRunning := string(*vals[idxSlaveSQLRunning].(*sql.RawBytes))
			channelName := "-"
			if idxChannelName >= 0 {
				channelName = string(*vals[idxChannelName].(*sql.RawBytes))
			}
			strSecondsBehindMaster := string(*vals[idxSecondsBehindMaster].(*sql.RawBytes))
			secondsBehindMaster, e := strconv.ParseInt(strSecondsBehindMaster, 10, 64)
			if e != nil {
				ch <- e
				return
			}

			status := 0
			if slaveIORunning != "Yes" || slaveSQLRunning != "Yes" {
				status = 2
			}
			if opts.Crit > 0 && secondsBehindMaster > opts.Crit {
				status = 2
			} else if opts.Warn > 0 && secondsBehindMaster > opts.Warn {
				status = 1
			}

			msg := fmt.Sprintf("%s=io:%s,sql:%s,behind:%d", channelName, slaveIORunning, slaveSQLRunning, secondsBehindMaster)
			switch status {
			case 0:
				okStatuses = append(okStatuses, msg)
			case 1:
				warnStatuses = append(warnStatuses, msg)
			case 2:
				critStatuses = append(critStatuses, msg)
			}
		}
		if err := rows.Err(); err != nil {
			ch <- err
			return
		}

		if i == 0 {
			ch <- fmt.Errorf("No replication settings")
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
