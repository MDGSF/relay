package relay

import (
	"io"

	"github.com/MDGSF/utils"
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
		return 0, err
	}

	bodylen := utils.BytesToInt32(headerBuf)
	bodyBuf := make([]byte, bodylen)
	readedBodyLen, err := io.ReadFull(r.Reader, bodyBuf)
	if err != nil || readedBodyLen != bodylen {
		return 0, err
	}

	plainBody, err := x.AesDecrypt(bodyBuf, r.cipher)
	if err != nil {
		return 0, err
	}
	p = plainBody
	return headerlen + bodylen, nil
}

type XWriter struct {
	io.Writer
	cipher []byte
}

func (w *XWriter) Write(p []byte) (n int, err error) {
	crypted, err := x.AesEncrypt(p, w.cipher)
	if err != nil {
		return 0, nil
	}

	pLen := len(crypted)
	n1, err := w.Writer.Write(utils.IntTo4Bytes(pLen))
	if err != nil {
		return 0, err
	}

	n2, err := w.Writer.Write(crypted)
	if err != nil {
		return 0, err
	}
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
