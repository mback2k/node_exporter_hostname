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

package compress

import (
	"compress/gzip"
	"io"
)

func NewGzipCompressor(r io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)
	go Stream(gw, r, func() {
		pw.Close()
	})
	return pr
}
