package common

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

func PrintRequestAndResponse(req *http.Request, rsp *http.Response) {
	dumpRequest(os.Stdout, req, rsp)
}

func dumpRequest(writer io.Writer, req *http.Request, rsp *http.Response) {
	type syncable interface {
		Sync() error
	}

	var reqStr string
	if req == nil {
		reqStr = "nil"
	} else {
		reqBody, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			return
		}

		reqStr = string(reqBody)
	}

	var rspStr string
	if rsp == nil {
		rspStr = "nil"
	} else {
		rspBody, err := httputil.DumpResponse(rsp, true)
		if err != nil {
			return
		}

		rspStr = string(rspBody)
	}

	_, _ = fmt.Fprintln(writer, ">>>>>>>> Request >>>>>>>>")
	_, _ = fmt.Fprintln(writer, reqStr)
	_, _ = fmt.Fprintln(writer, "<<<<<<<< Response <<<<<<<<")
	_, _ = fmt.Fprintln(writer, rspStr)

	if sync, ok := writer.(syncable); ok {
		_ = sync.Sync()
	}
}
