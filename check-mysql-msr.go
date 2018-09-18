package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
)

type mysqlSetting struct {
	Host string `short:"H" long:"host" default:"localhost" description:"Hostname"`
	Port string `short:"p" long:"port" default:"3306" description:"Port"`
	User string `short:"u" long:"user" default:"root" description:"Username"`
	Pass string `short:"P" long:"password" default:"" description:"Password"`
}

type connectionOpts struct {
	mysqlSetting
	Crit int64 `short:"c" long:"critical" description:"critical if seconds behind master is larger than this number"`
	Warn int64 `short:"w" long:"warning" description:"warning if seconds behind master is larger than this number"`
}

func main() {
	ckr := checkMsr()
	ckr.Name = "MySQL Multi Source Replication"
	ckr.Exit()
}

func checkMsr() *checkers.Checker {
	opts := connectionOpts{}
	psr := flags.NewParser(&opts, flags.Default)
	_, err := psr.Parse()
	if err != nil {
		os.Exit(1)
	}

	db := mysql.New("tcp", "", fmt.Sprintf("%s:%s", opts.mysqlSetting.Host, opts.mysqlSetting.Port), opts.mysqlSetting.User, opts.mysqlSetting.Pass, "")
	err = db.Connect()
	if err != nil {
		return checkers.Critical("couldn't connect DB")
	}
	defer db.Close()

	var okStatuses []string
	var warnStatuses []string
	var critStatuses []string
	rows, _, err := db.Query("SELECT CHANNEL_NAME FROM performance_schema.replication_connection_status")
	if err != nil {
		return checkers.Unknown(fmt.Sprintf("couldn't execute query: %s", err))
	}

	if len(rows) == 0 {
		return checkers.Unknown("no replication channels found")
	}

	for _, row := range rows {
		channelName := row.Str(0)
		query := fmt.Sprintf("SHOW SLAVE STATUS FOR CHANNEL '%s'", channelName)
		slaveRows, slaveRes, err := db.Query(query)
		if err != nil {
			return checkers.Unknown("couldn't execute query")
		}

		idxIoThreadRunning := slaveRes.Map("Slave_IO_Running")
		idxSQLThreadRunning := slaveRes.Map("Slave_SQL_Running")
		idxSecondsBehindMaster := slaveRes.Map("Seconds_Behind_Master")
		ioThreadStatus := slaveRows[0].Str(idxIoThreadRunning)
		sqlThreadStatus := slaveRows[0].Str(idxSQLThreadRunning)
		secondsBehindMaster := slaveRows[0].Int64(idxSecondsBehindMaster)

		st := 0
		if ioThreadStatus != "Yes" || sqlThreadStatus != "Yes" {
			st = 2
		}
		if opts.Crit > 0 && secondsBehindMaster > opts.Crit {
			st = 2
		} else if opts.Warn > 0 && secondsBehindMaster > opts.Warn {
			st = 1
		}
		msg := fmt.Sprintf("%s=io:%s,sql:%s,behind:%d", channelName, ioThreadStatus, sqlThreadStatus, secondsBehindMaster)
		if st == 0 {
			okStatuses = append(okStatuses, msg)
		} else if st == 1 {
			warnStatuses = append(warnStatuses, msg)
		} else if st == 2 {
			critStatuses = append(critStatuses, msg)
		}
	}

	var msgs []string
	if len(critStatuses) > 0 {
		msgs = append(msgs, "[C]"+strings.Join(critStatuses[0:], " "))
	}
	if len(warnStatuses) > 0 {
		msgs = append(msgs, "[W]"+strings.Join(warnStatuses[0:], " "))
	}
	if len(okStatuses) > 0 {
		msgs = append(msgs, "[O]"+strings.Join(okStatuses[0:], " "))
	}
	msg := strings.Join(msgs[0:], " | ")
	if len(critStatuses) > 0 {
		return checkers.Critical(msg)
	} else if len(warnStatuses) > 0 {
		return checkers.Warning(msg)
	}
	return checkers.Ok(msg)
}
