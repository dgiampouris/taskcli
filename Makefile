# Simple Makefile that builds the application with the default name.
# It also makes cleaning easier!

db := ${HOME}/.tasks.db
key := /dev/shm/.taskdb

files := $(strip $(foreach f,$(filenames),$(wildcard $(f))))

all:
	go build -o ${GOPATH}/bin/task /home/pxcel/go/src/github.com/dgiampouris/taskcli/main.go

.PHONY: clean
clean:
	rm -f $(db) $(key)

.PHONY: cleandb
cleandb:
	rm -f $(db)

.PHONY: cleankey
cleankey:
	rm -f $(key)
