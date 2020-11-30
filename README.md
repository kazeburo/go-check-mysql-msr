# check-mysql-msr

check MySQL Multi Source replication

# Usage

```
Usage:
  check-mysql-msr [OPTIONS]

Application Options:
      --defaults-extra-file= path to defaults-extra-file
      --mysql-socket=        path to mysql listen sock
  -H, --host=                Hostname (default: localhost)
  -p, --port=                Port (default: 3306)
  -u, --user=                Username (default: root)
  -P, --password=            Password
      --database=            database name connect to
      --timeout=             Timeout to connect mysql (default: 5s)
  -c, --critical=            critical if seconds behind master is larger than this number
  -w, --warning=             warning if seconds behind master is larger than this number
  -v, --version              Show version

Help Options:
  -h, --help                 Show this help message
```

Sample

```
$ ./check-mysql-msr
MySQL Multi Source Replication OK: [O]main-db=io:Yes,sql:Yes,behind:0 user-db=io:Yes,sql:Yes,behind:0 
```


## Install

Please download release page or `mkr plugin install kazeburo/go-check-mysql-msr`.