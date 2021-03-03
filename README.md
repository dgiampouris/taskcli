# Taskcli

A simple CLI task manager.

## Description

Taskcli is a very simple to-do list managing program meant to be used from the command line.
It uses a locally stored database to store all your tasks (or TODOs).
The database is encrypted when at rest and only gets decrypted when you interact with it.

## Install

Get the package with `go get` and build it with `go build`. Make sure that `$GOPATH/bin` is in your `$PATH`.

```
$ go get github.com/dgiampouris/taskcli
$ cd $GOPATH/src/github.com/dgiampouris/taskcli
$ go build -o $GOPATH/bin/task main.go
```

## Usage

Make sure to put any of the given tasks in double quotes (e.g. `"This is an example."`).

```
$ task add "TODO an example."
$ task list

Here's a list of your tasks:

1. TODO an example.
$ task delete 1

Here's a list of your tasks:

```
