# relay

## Go environment

First, you need golang environment to build the program.

https://golang.org/dl/

## Build

```
cd main
go build -o relay main.go
```

## Usage

relay is use to relay tcp stream, and it will encrypted data.

usage:

-faddr: front address, relay will listen at this address, and send data to backen address.
-baddr: backen address.
-frc: front cipher, use to encrypte data through front address.
-brc: backen cipher, use to encrypte data go through backen address.

## Delopyment

**your vps:** public ipv4 ip address : 10.10.10.1(this is fake ip, just use your vps ip address)

open sock5: **./sock5 -addr ":1082"**

open remote relay: **./relay -faddr ":1081" -baddr "10.10.10.1:1082" -frc "123"**

**your local pc:**

open local relay: **./relay -faddr ":1080" -baddr ":1081" -brc "123"**

open browser, **set proxy to localhost:1080**

browser will send data to local relay,

local relay will send data to remote relay,

remote relay will send data to sock5.

**what is sock5?** : go get github.com/MDGSF/sock5

## And you can use arbitrary numbers of relay.

If you have two vps, one is inner, the other is outter. Then you can relay your local pc to inner vps, and relay inner vps to outter vps.

**outter vps:** public ipv4 ip address : 10.10.10.2

open sock5: **./sock5 -addr ":1083"**

open remote relay: **./relay -faddr ":1082" -baddr "10.10.10.2:1083" -frc "123"**

**inner vps:** public ipv4 ip address : 10.10.10.1

open remote relay: **./relay -faddr ":1081" -baddr "10.10.10.2:1082"**

**your local pc:**

open local relay: **./relay -faddr ":1080" -baddr "10.10.10.1:1081" -brc "123"**

open browser, **set proxy to localhost:1080**


