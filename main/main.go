package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/MDGSF/relay"
	"github.com/MDGSF/utils/log"
)

var frontAddr = flag.String("faddr", "0.0.0.0:1080", "front address")
var backenAddr = flag.String("baddr", "", "backen address")
var frontCipher = flag.String("frc", "", "front cipher")
var backenCipher = flag.String("brc", "", "backen cipher")

func ioBridge(src io.Reader, dst io.Writer, shutdown chan bool) {
	defer func() {
		shutdown <- true
	}()

	buf := make([]byte, 8*1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			break
		}

		_, err = dst.Write(buf[:n])
		if err != nil {
			break
		}
	}
}

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

	var frontIO io.ReadWriter
	var backenIO io.ReadWriter
	if len(*frontCipher) > 0 {
		frontIO = relay.NewXReadWriter(frontconn, []byte(*frontCipher))
	} else {
		frontIO = frontconn
	}

	if len(*backenCipher) > 0 {
		backenIO = relay.NewXReadWriter(backconn, []byte(*backenCipher))
	} else {
		backenIO = backconn
	}

	shutdown := make(chan bool, 2)
	go ioBridge(frontIO, backenIO, shutdown)
	go ioBridge(backenIO, frontIO, shutdown)
	<-shutdown
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
