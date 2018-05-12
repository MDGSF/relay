package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/MDGSF/utils/log"
)

var frontAddr = flag.String("faddr", "0.0.0.0:1080", "front address")
var backenAddr = flag.String("baddr", "", "backen address")

var frontReadCipher = flag.String("frc", "", "front read cipher")
var frontWriteCipher = flag.String("fwc", "", "front write cipher")
var backenReadCipher = flag.String("brc", "", "backen read cipher")
var backenWriteCipher = flag.String("bwc", "", "backen write cipher")

func handleConnection(frontconn net.Conn) {
	frontNetwork := frontconn.RemoteAddr().Network()
	frontaddr := frontconn.RemoteAddr().String()
	log.Info("new frontconn = %v, %v", frontNetwork, frontaddr)
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic recover err = %v", err)
		}
		frontconn.Close()
		log.Info("close frontconn = %v, %v", frontNetwork, frontaddr)
	}()

	backconn, err := net.Dial("tcp", *backenAddr)
	if err != nil {
		log.Error("connect to back addr %v failed, err = %v", *backenAddr, err)
		return
	}

	backenNetwork := backconn.RemoteAddr().Network()
	backenaddr := backconn.RemoteAddr().String()
	log.Info("new backconn = %v, %v", backenNetwork, backenaddr)
	defer func() {
		backconn.Close()
		log.Info("close backconn = %v, %v", backenNetwork, backenaddr)
	}()
}

func main() {
	fmt.Println("relay startting...")
	flag.Parse()

	listener, err := net.Listen("tcp", *frontAddr)
	if err != nil {
		log.Error("listen %v failed, err = %v", *frontAddr, err)
		return
	}
	log.Info("listen on %v...\n", *frontAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}
