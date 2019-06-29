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

package hostmetrics

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type HostMetricsReader struct {
	Source io.ReadCloser

	reader *io.PipeReader
	writer *io.PipeWriter

	scanner *bufio.Scanner
	printer *bufio.Writer

	hostnameLabel string
}

func NewHostMetricsReader(r io.ReadCloser) *HostMetricsReader {
	return &HostMetricsReader{Source: r}
}

func (r *HostMetricsReader) Read(p []byte) (int, error) {
	if r.reader == nil {
		r.readLines()
	}
	return r.reader.Read(p)
}

func (r *HostMetricsReader) Close() error {
	if r.reader != nil {
		r.reader.Close()
	}
	return r.Source.Close()
}

func (r *HostMetricsReader) getScanner() *bufio.Scanner {
	if r.scanner == nil {
		r.scanner = bufio.NewScanner(r.Source)
	}
	return r.scanner
}

func (r *HostMetricsReader) getPrinter() *bufio.Writer {
	if r.printer == nil {
		r.printer = bufio.NewWriter(r.writer)
	}
	return r.printer
}

func (r *HostMetricsReader) getHostnameLabel() string {
	if r.hostnameLabel == "" {
		var b strings.Builder
		hostname, err := os.Hostname()
		b.Grow(11 + len(hostname))
		b.WriteString("hostname=")
		b.WriteRune('"')
		if err == nil {
			b.WriteString(hostname)
		}
		b.WriteRune('"')
		r.hostnameLabel = b.String()
	}
	return r.hostnameLabel
}

func (r *HostMetricsReader) readLines() {
	if r.reader == nil && r.writer == nil {
		r.reader, r.writer = io.Pipe()
		go r.streamLines()
	}
}

func (r *HostMetricsReader) streamLines() {
	scanner := r.getScanner()
	printer := r.getPrinter()
	for scanner.Scan() {
		line := scanner.Text()
		line = r.modifyLine(line)
		if _, err := printer.WriteString(line); err != nil {
			r.writer.CloseWithError(err)
			return
		}
		if _, err := printer.WriteRune('\n'); err != nil {
			r.writer.CloseWithError(err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		r.writer.CloseWithError(err)
		return
	}
	err := printer.Flush()
	r.writer.CloseWithError(err)
}

func (r *HostMetricsReader) modifyLine(line string) string {
	if line == "" {
		return line
	}
	if line[0] == '#' {
		return line
	}

	var b strings.Builder
	lbl := r.getHostnameLabel()
	if pos := strings.IndexRune(line, '{'); pos >= 0 {
		pos++
		b.Grow(1 + len(line) + len(lbl))
		b.WriteString(line[:pos])
		b.WriteString(lbl)
		b.WriteRune(',')
		b.WriteString(line[pos:])
	} else if pos = strings.IndexRune(line, ' '); pos >= 0 {
		b.Grow(2 + len(line) + len(lbl))
		b.WriteString(line[:pos])
		b.WriteRune('{')
		b.WriteString(lbl)
		b.WriteRune('}')
		b.WriteString(line[pos:])
	} else {
		return line
	}
	return b.String()
}
