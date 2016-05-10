# go-check-mysql-msr

```
$ ./check-mysql-msr
MySQL Multi Source Replication OK: [O]main-db=io:Yes,sql:Yes,behind:0 user-db=io:Yes,sql:Yes,behind:0 
```

usage

```
$ ./check-mysql-msr -h
Usage:
  check-mysql-msr [OPTIONS]

Application Options:
  -H, --host=     Hostname (localhost)
  -p, --port=     Port (3306)
  -u, --user=     Username (root)
  -P, --password= Password
  -c, --critical= critical if seconds behind master is larger than this number
  -w, --warning=  warning if seconds behind master is larger than this number

Help Options:
  -h, --help      Show this help message

```


