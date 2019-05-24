/*
	node_exporter_hostname - A prom node_exporter proxy with hostname labels.
	Copyright (C) 2019  Marc Hoersken <info@marc-hoersken.de>

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"

	"github.com/mback2k/node_exporter_hostname/compress"
	"github.com/mback2k/node_exporter_hostname/hostmetrics"
)

func prom() {
	cmd := exec.Command("/bin/node_exporter", os.Args[1:]...)
	log.Fatal(cmd.Run())
}

func main() {
	go prom()
	url, _ := url.ParseRequestURI("http://localhost:9100/")
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ModifyResponse = modifyResponse
	http.Handle("/", proxy)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func modifyResponse(r *http.Response) error {
	body := r.Body
	encoding := r.Header.Get("Content-Encoding")
	switch encoding {
	case "gzip":
		gr, err := gzip.NewReader(body)
		if err != nil {
			return err
		}
		body = gr
	}
	body = hostmetrics.NewHostMetricsReader(body)
	switch encoding {
	case "gzip":
		body = compress.NewGzipCompressor(body)
	}
	r.Body = body
	return nil
}
