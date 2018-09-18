VERSION=0.0.5

all: check-mysql-msr

.PHONY: check-mysql-msr

bundle:
	dep ensure

update:
	dep ensure -update

check-mysql-msr: check-mysql-msr.go
	go build $(LDFLAGS) -o check-mysql-msr

linux: check-mysql-msr.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o check-mysql-msr

fmt:
	go fmt ./...

clean:
	rm -rf check-mysql-msr check-mysql-msr-*.tar.gz

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
