// The MIT License (MIT)
//
// Copyright (c) 2015 xtaci
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package kcp

import (
	"sync/atomic"

	"github.com/pkg/errors"

	UT "github.com/hadi77ir/go-udp/types"
)

// defaultTx is the default procedure to transmit data
func (s *UDPSession) tx(tx [][]byte) {
	dataN := 0
	pktN := 0
	for _, pkt := range tx {
		if n, _, err := s.conn.Write(pkt, []byte{}, 0, UT.ECNUnsupported); err != nil {
			s.notifyWriteError(errors.WithStack(err))
		} else {
			pktN++
			dataN += n
		}
	}
	atomic.AddUint64(&DefaultSnmp.OutPkts, uint64(pktN))
	atomic.AddUint64(&DefaultSnmp.OutBytes, uint64(dataN))
}
