# gows

## Overview

...

## Install

`go get github.com/sung1011/gows`

## Getting Started

### ping

server `go run demo/s/main.go`  
client `go run demo/c/main.go 4 ping`  // get cost time

### echo

server `go run demo/s/main.go`  
client `go run demo/c/main.go 4 echo foo`  // request and response
client `go run demo/c/main.go 4 echo bar`  // request and response

### push

server `go run demo/s/main.go`  
client1 `go run demo/c/main.go 1 push foo 2`  // uid 1 push msg to 2
client2 `go run demo/c/main.go 2 push bar 1`  // uid 2 push msg to 1
client3 `go run demo/c/main.go 3 push baz`  // push all

## Doc

### proto

text  
json  
crypt  
...

## License

...
