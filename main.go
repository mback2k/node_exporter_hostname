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
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/mback2k/node_exporter_hostname/compress"
	"github.com/mback2k/node_exporter_hostname/hostmetrics"
)

const (
	defaultListenAddress = ":8080"
	defaultScrapeMetrics = "http://localhost:9100/"
	defaultLaunchProgram = "/bin/node_exporter"
)

var (
	listenAddress = flag.String("listen-address", defaultListenAddress, "The address to listen on for HTTP requests.")
	scrapeMetrics = flag.String("scrape-metrics", defaultScrapeMetrics, "The URL to proxy to retrieve and serve metrics.")
	launchProgram = flag.String("launch-program", defaultLaunchProgram, "The command to run to retrieve and serve metrics.")
	updateMetrics = flag.Bool("modify-metrics", true, "Enable or disable modification of the scraped metrics.")
)

func main() {
	flag.Parse()
	ok, err := strconv.ParseBool(*launchProgram)
	if err == nil && ok {
		go launchMetrics(defaultLaunchProgram)
	} else if err != nil && *launchProgram != "" {
		go launchMetrics(*launchProgram)
	}

	url, _ := url.ParseRequestURI(*scrapeMetrics)
	proxy := httputil.NewSingleHostReverseProxy(url)
	if *updateMetrics {
		proxy.ModifyResponse = modifyMetrics
	}
	http.Handle("/", proxy)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func launchMetrics(p string) {
	cmd := exec.Command(p, flag.Args()...)
	log.Fatal(cmd.Run())
}

func modifyMetrics(r *http.Response) error {
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
	r.Header.Del("Content-Length")
	r.ContentLength = -1
	r.Body = body
	return nil
}
