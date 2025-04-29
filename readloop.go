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
	"github.com/pkg/errors"
)

// defaultReadLoop is the standard procedure for reading from a connection
func (s *UDPSession) readLoop() {
	buf := make([]byte, mtuLimit)
	for {
		if n, _, _, err := s.conn.Read(buf, []byte{}); err == nil {
			if s.isClosed() {
				return
			}
			s.packetInput(buf[:n])
		} else {
			s.notifyReadError(errors.WithStack(err))
			return
		}
	}
}

// defaultReadLoop is the standard procedure for reading and accepting connections on a listener
func (l *Listener) monitor() {
	for {
		conn, err := l.conn.Accept()
		if err != nil {
			l.notifyReadError(errors.WithStack(err))
			return
		}
		go func() {
			buf := make([]byte, mtuLimit)
			for {
				if err := l.packetInput(buf, conn); err != nil {
					_ = conn.Close()
					l.notifyReadError(errors.WithStack(err))
					return
				}
			}
		}()
	}
}
