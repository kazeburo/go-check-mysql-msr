VERSION=0.0.3

all: check-mysql-msr

.PHONY: check-mysql-msr

gom:
	go get -u github.com/mattn/gom

bundle:
	gom install

check-mysql-msr: check-mysql-msr.go
	gom build -o check-mysql-msr

linux: check-mysql-msr.go
	GOOS=linux GOARCH=amd64 gom build -o check-mysql-msr

fmt:
	go fmt ./...

dist:
	git archive --format tgz HEAD -o check-mysql-msr-$(VERSION).tar.gz --prefix check-mysql-msr-$(VERSION)/

clean:
	rm -rf check-mysql-msr check-mysql-msr-*.tar.gz

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
