package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"hash"
	"io"
	"net"
	"time"

	"github.com/MDGSF/utils"
	"github.com/MDGSF/utils/log"
	"github.com/MDGSF/utils/x"
)

var frontAddr = flag.String("faddr", "0.0.0.0:1080", "front address")
var backenAddr = flag.String("baddr", "", "backen address")
var frontCipher = flag.String("frc", "", "front cipher")
var backenCipher = flag.String("brc", "", "backen cipher")
var frontKey []byte
var backenKey []byte

const (
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
)

func padCipherTo32Key(cipher string) []byte {
	if len(cipher) == 0 {
		return nil
	}

	var hasher hash.Hash
	hasher = sha1.New()
	hasher.Write([]byte(cipher))
	digest := hex.EncodeToString(hasher.Sum(nil))
	digestByteArr := []byte(digest)
	return digestByteArr[:32]
}

type TXConn struct {
	conn net.Conn
	key  []byte
}

func NewXConn(conn net.Conn, key []byte) *TXConn {
	return &TXConn{
		conn: conn,
		key:  key,
	}
}

func (c *TXConn) Read(p []byte) (n int, err error) {
	if c.key == nil || len(c.key) == 0 {
		return c.conn.Read(p)
	}
	return c.xread(p)
}

func (c *TXConn) xread(p []byte) (n int, err error) {
	err = c.conn.SetReadDeadline(time.Now().Add(DefaultReadTimeout))
	if err != nil {
		log.Printf("SetReadDeadline failed, err = %v\n", err)
		return 0, err
	}

	headerBuf := make([]byte, 4)
	headerlen, err := c.conn.Read(headerBuf)
	if err != nil || headerlen != 4 {
		log.Error("x read failed, headerlen = %v, err = %v", headerlen, err)
		return 0, err
	}

	bodylen := utils.BytesToInt32(headerBuf)
	bodyBuf := make([]byte, bodylen)
	readedBodyLen, err := io.ReadFull(c.conn, bodyBuf)
	if err != nil || readedBodyLen != bodylen {
		log.Error("read body failed, readedbodylen = %v, bodylen = %v, err = %v", readedBodyLen, bodylen, err)
		return 0, err
	}

	plainBody, err := x.AesDecrypt(bodyBuf, c.key)
	if err != nil {
		log.Error("aes decrypt failed, err = %v", err)
		return 0, err
	}
	copy(p, plainBody)

	log.Verbose("Read c.key = %v,  headerlen = %v, bodylen = %v, len(p) = %v, len(plainBody) = %v, plainBody = %v",
		c.key, headerlen, bodylen, len(p), len(plainBody), plainBody)

	return len(plainBody), nil
}

func (c *TXConn) Write(p []byte) (n int, err error) {
	if c.key == nil || len(c.key) == 0 {
		return c.conn.Write(p)
	}
	return c.xwrite(p)
}

func (c *TXConn) xwrite(p []byte) (n int, err error) {
	log.Verbose("Write len(p) = %v, c.key = %v, p = %v", len(p), c.key, p)

	var crypted []byte
	crypted, err = x.AesEncrypt(p, c.key)
	if err != nil {
		log.Error("aes encrypt failed, err = %v", err)
		return 0, err
	}

	err = c.conn.SetWriteDeadline(time.Now().Add(DefaultWriteTimeout))
	if err != nil {
		log.Printf("SetWriteDeadline failed, err = %v\n", err)
		return 0, err
	}

	pLen := len(crypted)
	n1, err := c.conn.Write(utils.IntTo4Bytes(pLen))
	if err != nil {
		log.Error("x write failed, err = %v", err)
		return 0, err
	}

	n2, err := c.conn.Write(crypted)
	if err != nil {
		log.Error("x write failed, err = %v", err)
		return 0, err
	}

	log.Verbose("x write n1 = %v, n2 = %v, pLen = %v, crypted = %v", n1, n2, pLen, crypted)
	return n1 + n2, nil
}

func ioBridge(src *TXConn, dst *TXConn) {
	buf := utils.GLeakyBuf.Get()
	defer utils.GLeakyBuf.Put(buf)
	for {
		n, err := src.Read(buf)
		if err != nil || n == 0 {
			log.Error("read failed, err = %v", err)
			return
		}

		_, err = dst.Write(buf[:n])
		if err != nil {
			log.Error("write failed, err = %v", err)
			return
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

	frontIO := NewXConn(frontconn, frontKey)
	backenIO := NewXConn(backconn, backenKey)
	go ioBridge(frontIO, backenIO)
	ioBridge(backenIO, frontIO)
}

func main() {
	flag.Parse()
	log.SetLevel(log.InfoLevel)

	frontKey = padCipherTo32Key(*frontCipher)
	backenKey = padCipherTo32Key(*backenCipher)

	listener, err := net.Listen("tcp", *frontAddr)
	if err != nil {
		log.Error("listen %v failed, err = %v", *frontAddr, err)
		return
	}
	log.Info("listen on %v...\n", *frontAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("accept failed, err = %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
