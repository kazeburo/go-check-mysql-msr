module github.com/kazeburo/go-check-mysql-msr

go 1.15

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/kazeburo/go-mysqlflags v0.0.8
	github.com/mackerelio/checkers v0.0.0-20200428063449-52cfb2c2c52c
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/shirou/gopsutil v3.20.11+incompatible // indirect
	golang.org/x/sys v0.0.0-20201202213521-69691e467435 // indirect
	golang.org/x/tools v0.0.0-20201202200335-bef1c476418a // indirect
)

replace github.com/mitchellh/mapstructure => github.com/kazeburo/mapstructure v1.4.1-0.20201203061123-1b85cddd5215
