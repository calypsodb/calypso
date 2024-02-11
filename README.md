# calypso
Dead simple KV store in Go

# Building and Usage

## Build

* Using the Makefile

```
make
```

* Using the Go compiler

```sh
go build -o calypso seagull/calypso/*
go build -o calypsod seagull/calypsod/*
```
## Use

`calypsod` is the Daemon

`calypso` is the CLI tool


With the `calypso` CLI you can interact with a CalypsoDB Bitcask directory as:

```
./calypso put x 10

./calypso get x


```