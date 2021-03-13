# Simple Makefile that builds the application with the default name.
# It also makes cleaning easier!

db := ${HOME}/.tasks.db
key := /dev/shm/.taskdb

all:
	go build -o ${GOPATH}/bin/task ${GOPATH}/src/github.com/dgiampouris/taskcli/main.go

.PHONY: clean
clean:
	rm -f $(db) $(key)

.PHONY: cleandb
cleandb:
	rm -f $(db)

.PHONY: cleankey
cleankey:
	rm -f $(key)
