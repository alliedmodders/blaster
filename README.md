blaster
=======

Source and GoldSrc Query Tool

Building
--------

1. Make sure you have Golang installed, (see: http://golang.org/)
2. Make sure your Go environment is set up. Example:
```
export GOROOT=~/tools/go
export GOPATH=~/go
export PATH="$PATH:$GOROOT/bin:$GOPATH/bin"
```
3. Get the source code and its dependencies:
```
go get github.com/alliedmodders/blaster
```
4. Build:
```
go install
```
5. The `blaster` binary wll be in `$GOPATH/bin/`.
