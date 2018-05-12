package relay

import (
	"io"

	"github.com/MDGSF/utils"
	"github.com/MDGSF/utils/log"
	"github.com/MDGSF/utils/x"
)

type XReader struct {
	io.Reader
	cipher []byte
}

func (r *XReader) Read(p []byte) (n int, err error) {
	headerBuf := make([]byte, 4)
	headerlen, err := r.Reader.Read(headerBuf)
	if err != nil || headerlen != 4 {
		log.Error("x read failed, headerlen = %v, err = %v", headerlen, err)
		return 0, err
	}

	bodylen := utils.BytesToInt32(headerBuf)
	bodyBuf := make([]byte, bodylen)
	readedBodyLen, err := io.ReadFull(r.Reader, bodyBuf)
	if err != nil || readedBodyLen != bodylen {
		log.Error("read body failed, readedbodylen = %v, bodylen = %v, err = %v", readedBodyLen, bodylen, err)
		return 0, err
	}

	plainBody, err := x.AesDecrypt(bodyBuf, r.cipher)
	if err != nil {
		log.Error("aes decrypt failed, err = %v", err)
		return 0, err
	}
	copy(p, plainBody)

	log.Verbose("Read r.cipher = %v,  headerlen = %v, bodylen = %v, len(p) = %v, len(plainBody) = %v, plainBody = %v",
		r.cipher, headerlen, bodylen, len(p), len(plainBody), plainBody)

	return len(plainBody), nil
}

type XWriter struct {
	io.Writer
	cipher []byte
}

func (w *XWriter) Write(p []byte) (n int, err error) {
	log.Verbose("Write len(p) = %v, w.cipher = %v, p = %v", len(p), w.cipher, p)
	crypted, err := x.AesEncrypt(p, w.cipher)
	if err != nil {
		log.Error("aed encrypt failed, err = %v", err)
		return 0, nil
	}

	pLen := len(crypted)
	n1, err := w.Writer.Write(utils.IntTo4Bytes(pLen))
	if err != nil {
		log.Error("x write failed, err = %v", err)
		return 0, err
	}

	n2, err := w.Writer.Write(crypted)
	if err != nil {
		log.Error("x write failed, err = %v", err)
		return 0, err
	}

	log.Verbose("x write n1 = %v, n2 = %v, pLen = %v, crypted = %v", n1, n2, pLen, crypted)

	return n1 + n2, nil
}

type XReadWriter struct {
	*XReader
	*XWriter
}

func NewXReadWriter(r io.ReadWriter, cipher []byte) XReadWriter {
	xr := XReader{Reader: r, cipher: cipher}
	xw := XWriter{Writer: r, cipher: cipher}
	return XReadWriter{&xr, &xw}
}
