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
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"

	U "github.com/hadi77ir/go-udp"
	UT "github.com/hadi77ir/go-udp/types"
)

var baseport = uint32(10000)
var key = []byte("testkey")
var pass = pbkdf2.Key(key, []byte("testsalt"), 4096, 32, sha1.New)

func init() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.Println("beginning tests, encryption:salsa20, fec:10/3")
}

func dialEcho(port int) (*UDPSession, error) {
	//block, _ := NewNoneBlockCrypt(pass)
	//block, _ := NewSimpleXORBlockCrypt(pass)
	//block, _ := NewTEABlockCrypt(pass[:16])
	//block, _ := NewAESBlockCrypt(pass)
	block, _ := NewSalsa20BlockCrypt(pass)
	sess, err := DialWithOptions(fmt.Sprintf("127.0.0.1:%v", port), block, 10, 3)
	if err != nil {
		panic(err)
	}

	sess.SetStreamMode(true)
	sess.SetStreamMode(false)
	sess.SetStreamMode(true)
	sess.SetWindowSize(1024, 1024)
	sess.SetReadBuffer(16 * 1024 * 1024)
	sess.SetWriteBuffer(16 * 1024 * 1024)
	sess.SetStreamMode(true)
	sess.SetNoDelay(1, 10, 2, 1)
	sess.SetMtu(1400)
	sess.SetMtu(1600)
	sess.SetMtu(1400)
	sess.SetACKNoDelay(true)
	sess.SetACKNoDelay(false)
	sess.SetDeadline(time.Now().Add(time.Minute))
	return sess, err
}

func dialSink(port int) (*UDPSession, error) {
	sess, err := DialWithOptions(fmt.Sprintf("127.0.0.1:%v", port), nil, 0, 0)
	if err != nil {
		panic(err)
	}

	sess.SetStreamMode(true)
	sess.SetWindowSize(1024, 1024)
	sess.SetReadBuffer(16 * 1024 * 1024)
	sess.SetWriteBuffer(16 * 1024 * 1024)
	sess.SetStreamMode(true)
	sess.SetNoDelay(1, 10, 2, 1)
	sess.SetMtu(1400)
	sess.SetACKNoDelay(false)
	sess.SetDeadline(time.Now().Add(time.Minute))
	return sess, err
}

func dialTinyBufferEcho(port int) (*UDPSession, error) {
	//block, _ := NewNoneBlockCrypt(pass)
	//block, _ := NewSimpleXORBlockCrypt(pass)
	//block, _ := NewTEABlockCrypt(pass[:16])
	//block, _ := NewAESBlockCrypt(pass)
	block, _ := NewSalsa20BlockCrypt(pass)
	sess, err := DialWithOptions(fmt.Sprintf("127.0.0.1:%v", port), block, 10, 3)
	if err != nil {
		panic(err)
	}
	return sess, err
}

// ////////////////////////
func listenEcho(port int) (net.Listener, error) {
	//block, _ := NewNoneBlockCrypt(pass)
	//block, _ := NewSimpleXORBlockCrypt(pass)
	//block, _ := NewTEABlockCrypt(pass[:16])
	//block, _ := NewAESBlockCrypt(pass)
	block, _ := NewSalsa20BlockCrypt(pass)
	return ListenWithOptions(fmt.Sprintf("127.0.0.1:%v", port), block, 10, 1)
}
func listenTinyBufferEcho(port int) (net.Listener, error) {
	//block, _ := NewNoneBlockCrypt(pass)
	//block, _ := NewSimpleXORBlockCrypt(pass)
	//block, _ := NewTEABlockCrypt(pass[:16])
	//block, _ := NewAESBlockCrypt(pass)
	block, _ := NewSalsa20BlockCrypt(pass)
	return ListenWithOptions(fmt.Sprintf("127.0.0.1:%v", port), block, 10, 3)
}

func listenSink(port int) (net.Listener, error) {
	return ListenWithOptions(fmt.Sprintf("127.0.0.1:%v", port), nil, 0, 0)
}

func echoServer(port int) net.Listener {
	l, err := listenEcho(port)
	if err != nil {
		panic(err)
	}

	go func() {
		kcplistener := l.(*Listener)
		kcplistener.SetReadBuffer(4 * 1024 * 1024)
		kcplistener.SetWriteBuffer(4 * 1024 * 1024)
		kcplistener.SetDSCP(46)
		for {
			s, err := l.Accept()
			if err != nil {
				return
			}

			// coverage test
			s.(*UDPSession).SetReadBuffer(4 * 1024 * 1024)
			s.(*UDPSession).SetWriteBuffer(4 * 1024 * 1024)
			go handleEcho(s.(*UDPSession))
		}
	}()

	return l
}

func sinkServer(port int) net.Listener {
	l, err := listenSink(port)
	if err != nil {
		panic(err)
	}

	go func() {
		kcplistener := l.(*Listener)
		kcplistener.SetReadBuffer(4 * 1024 * 1024)
		kcplistener.SetWriteBuffer(4 * 1024 * 1024)
		kcplistener.SetDSCP(46)
		for {
			s, err := l.Accept()
			if err != nil {
				return
			}

			go handleSink(s.(*UDPSession))
		}
	}()

	return l
}

func tinyBufferEchoServer(port int) net.Listener {
	l, err := listenTinyBufferEcho(port)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			s, err := l.Accept()
			if err != nil {
				return
			}
			go handleTinyBufferEcho(s.(*UDPSession))
		}
	}()
	return l
}

///////////////////////////

func handleEcho(conn *UDPSession) {
	conn.SetStreamMode(true)
	conn.SetWindowSize(4096, 4096)
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetDSCP(46)
	conn.SetMtu(1400)
	conn.SetACKNoDelay(false)
	conn.SetReadDeadline(time.Now().Add(time.Hour))
	conn.SetWriteDeadline(time.Now().Add(time.Hour))
	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}

func handleSink(conn *UDPSession) {
	conn.SetStreamMode(true)
	conn.SetWindowSize(4096, 4096)
	conn.SetNoDelay(1, 10, 2, 1)
	conn.SetDSCP(46)
	conn.SetMtu(1400)
	conn.SetACKNoDelay(false)
	conn.SetReadDeadline(time.Now().Add(time.Hour))
	conn.SetWriteDeadline(time.Now().Add(time.Hour))
	buf := make([]byte, 65536)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
	}
}

func handleTinyBufferEcho(conn *UDPSession) {
	conn.SetStreamMode(true)
	buf := make([]byte, 2)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		conn.Write(buf[:n])
	}
}

///////////////////////////

func TestTimeout(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 10)

	//timeout
	cli.SetDeadline(time.Now().Add(time.Second))
	<-time.After(2 * time.Second)
	n, err := cli.Read(buf)
	if n != 0 || err == nil {
		t.Fail()
	}
	cli.Close()
}

func TestSendRecv(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}
	cli.SetWriteDelay(true)
	cli.SetDUP(1)
	const N = 100
	buf := make([]byte, 10)
	for i := 0; i < N; i++ {
		msg := fmt.Sprintf("hello%v", i)
		cli.Write([]byte(msg))
		if n, err := cli.Read(buf); err == nil {
			if string(buf[:n]) != msg {
				t.Fail()
			}
		} else {
			panic(err)
		}
	}
	cli.Close()
}

func TestSendVector(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}
	cli.SetWriteDelay(false)
	const N = 100
	buf := make([]byte, 20)
	v := make([][]byte, 2)
	for i := 0; i < N; i++ {
		v[0] = []byte(fmt.Sprintf("hello%v", i))
		v[1] = []byte(fmt.Sprintf("world%v", i))
		msg := fmt.Sprintf("hello%vworld%v", i, i)
		cli.WriteBuffers(v)
		if n, err := cli.Read(buf); err == nil {
			if string(buf[:n]) != msg {
				t.Error(string(buf[:n]), msg)
			}
		} else {
			panic(err)
		}
	}
	cli.Close()
}

func TestTinyBufferReceiver(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := tinyBufferEchoServer(port)
	defer l.Close()

	cli, err := dialTinyBufferEcho(port)
	if err != nil {
		panic(err)
	}
	const N = 100
	snd := byte(0)
	fillBuffer := func(buf []byte) {
		for i := 0; i < len(buf); i++ {
			buf[i] = snd
			snd++
		}
	}

	rcv := byte(0)
	check := func(buf []byte) bool {
		for i := 0; i < len(buf); i++ {
			if buf[i] != rcv {
				return false
			}
			rcv++
		}
		return true
	}
	sndbuf := make([]byte, 7)
	rcvbuf := make([]byte, 7)
	for i := 0; i < N; i++ {
		fillBuffer(sndbuf)
		cli.Write(sndbuf)
		if n, err := io.ReadFull(cli, rcvbuf); err == nil {
			if !check(rcvbuf[:n]) {
				t.Fail()
			}
		} else {
			panic(err)
		}
	}
	cli.Close()
}

func TestClose(t *testing.T) {
	var n int
	var err error

	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}

	// double close
	cli.Close()
	if cli.Close() == nil {
		t.Fatal("double close misbehavior")
	}

	// write after close
	buf := make([]byte, 10)
	n, err = cli.Write(buf)
	if n != 0 || err == nil {
		t.Fatal("write after close misbehavior")
	}

	// write, close, read, read
	cli, err = dialEcho(port)
	if err != nil {
		panic(err)
	}
	if n, err = cli.Write(buf); err != nil {
		t.Fatal("write misbehavior")
	}

	// wait until data arrival
	time.Sleep(2 * time.Second)
	// drain
	cli.Close()
	n, err = io.ReadFull(cli, buf)
	if err != nil {
		t.Fatal("closed conn drain bytes failed", err, n)
	}

	// after drain, read should return error
	n, err = cli.Read(buf)
	if n != 0 || err == nil {
		t.Fatal("write->close->drain->read misbehavior", err, n)
	}
	cli.Close()
}

func TestParallel1024CLIENT_64BMSG_64CNT(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	var wg sync.WaitGroup
	wg.Add(1024)
	for i := 0; i < 1024; i++ {
		go parallel_client(&wg, port)
	}
	wg.Wait()
}

func parallel_client(wg *sync.WaitGroup, port int) (err error) {
	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}

	err = echo_tester(cli, 64, 64)
	cli.Close()
	wg.Done()
	return
}

func BenchmarkEchoSpeed4K(b *testing.B) {
	speedclient(b, 4096)
}

func BenchmarkEchoSpeed64K(b *testing.B) {
	speedclient(b, 65536)
}

func BenchmarkEchoSpeed512K(b *testing.B) {
	speedclient(b, 524288)
}

func BenchmarkEchoSpeed1M(b *testing.B) {
	speedclient(b, 1048576)
}

func speedclient(b *testing.B, nbytes int) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	b.ReportAllocs()
	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}

	if err := echo_tester(cli, nbytes, b.N); err != nil {
		b.Fail()
	}
	b.SetBytes(int64(nbytes))
	cli.Close()
}

func BenchmarkSinkSpeed4K(b *testing.B) {
	sinkclient(b, 4096)
}

func BenchmarkSinkSpeed64K(b *testing.B) {
	sinkclient(b, 65536)
}

func BenchmarkSinkSpeed256K(b *testing.B) {
	sinkclient(b, 524288)
}

func BenchmarkSinkSpeed1M(b *testing.B) {
	sinkclient(b, 1048576)
}

func sinkclient(b *testing.B, nbytes int) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := sinkServer(port)
	defer l.Close()

	b.ReportAllocs()
	cli, err := dialSink(port)
	if err != nil {
		panic(err)
	}

	sink_tester(cli, nbytes, b.N)
	b.SetBytes(int64(nbytes))
	cli.Close()
}

func echo_tester(cli net.Conn, msglen, msgcount int) error {
	buf := make([]byte, msglen)
	for i := 0; i < msgcount; i++ {
		// send packet
		if _, err := cli.Write(buf); err != nil {
			return err
		}

		// receive packet
		nrecv := 0
		for {
			n, err := cli.Read(buf)
			if err != nil {
				return err
			} else {
				nrecv += n
				if nrecv == msglen {
					break
				}
			}
		}
	}
	return nil
}

func sink_tester(cli *UDPSession, msglen, msgcount int) error {
	// sender
	buf := make([]byte, msglen)
	for i := 0; i < msgcount; i++ {
		if _, err := cli.Write(buf); err != nil {
			return err
		}
	}
	return nil
}

func TestSNMP(t *testing.T) {
	t.Log(DefaultSnmp.Copy())
	t.Log(DefaultSnmp.Header())
	t.Log(DefaultSnmp.ToSlice())
	DefaultSnmp.Reset()
	t.Log(DefaultSnmp.ToSlice())
}

func TestListenerClose(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l, err := ListenWithOptions(fmt.Sprintf("127.0.0.1:%v", port), nil, 10, 3)
	if err != nil {
		t.Fail()
	}
	l.SetReadDeadline(time.Now().Add(time.Second))
	l.SetWriteDeadline(time.Now().Add(time.Second))
	l.SetDeadline(time.Now().Add(time.Second))
	time.Sleep(2 * time.Second)
	if _, err := l.Accept(); err == nil {
		t.Fail()
	}

	l.Close()
	fakeaddr, _ := net.ResolveUDPAddr("udp6", "127.0.0.1:1111")
	if l.closeSession(fakeaddr) {
		t.Fail()
	}
}

// A wrapper for net.PacketConn that remembers when Close has been called.
type closedFlagSuperConn struct {
	UT.SuperConn
	Closed bool
}

func (c *closedFlagSuperConn) Close() error {
	c.Closed = true
	return c.SuperConn.Close()
}

func newClosedFlagSuperConn(c UT.SuperConn) *closedFlagSuperConn {
	return &closedFlagSuperConn{c, false}
}

// A wrapper for net.PacketConn that remembers when Close has been called.
type closedFlagPacketConn struct {
	UT.PacketConn
	Closed bool
}

func (c *closedFlagPacketConn) Close() error {
	c.Closed = true
	return c.PacketConn.Close()
}

func newClosedFlagPacketConn(c UT.PacketConn) *closedFlagPacketConn {
	return &closedFlagPacketConn{c, false}
}

// Listener should not close a net.PacketConn that it did not create.
// https://github.com/xtaci/kcp-go/issues/165
func TestListenerNonOwnedPacketConn(t *testing.T) {
	// Create a net.PacketConn not owned by the Listener.
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	c, err := U.Listen("udp", addr)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// Make it remember when it has been closed.
	pconn := newClosedFlagSuperConn(c)

	l, err := ServeConn(nil, 0, 0, pconn)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	if pconn.Closed {
		t.Fatal("non-owned PacketConn closed before Listener.Close()")
	}

	err = l.Close()
	if err != nil {
		panic(err)
	}

	if pconn.Closed {
		t.Fatal("non-owned PacketConn closed after Listener.Close()")
	}
}

// UDPSession should not close a net.PacketConn that it did not create.
// https://github.com/xtaci/kcp-go/issues/165
func TestUDPSessionNonOwnedPacketConn(t *testing.T) {
	l := sinkServer(17890)
	defer l.Close()

	// Create a net.PacketConn not owned by the UDPSession.
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:17890")
	if err != nil {
		panic(err)
	}
	c, err := U.Dial("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// Make it remember when it has been closed.
	pconn := newClosedFlagPacketConn(c)

	client, err := NewConn2(l.Addr(), nil, 0, 0, pconn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if pconn.Closed {
		t.Fatal("non-owned PacketConn closed before UDPSession.Close()")
	}

	err = client.Close()
	if err != nil {
		panic(err)
	}

	if pconn.Closed {
		t.Fatal("non-owned PacketConn closed after UDPSession.Close()")
	}
}

// this function test the data correctness with FEC and encryption enabled
func TestReliability(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	l := echoServer(port)
	defer l.Close()

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}
	cli.SetWriteDelay(false)
	const N = 100000
	buf := make([]byte, 128)
	msg := make([]byte, 128)

	for i := 0; i < N; i++ {
		io.ReadFull(rand.Reader, msg)
		cli.Write([]byte(msg))
		if n, err := io.ReadFull(cli, buf); err == nil {
			if !bytes.Equal(buf[:n], msg) {
				t.Fail()
			}
		} else {
			panic(err)
		}
	}
	cli.Close()
}

func TestControl(t *testing.T) {
	port := int(atomic.AddUint32(&baseport, 1))
	block, _ := NewSalsa20BlockCrypt(pass)
	l, err := ListenWithOptions(fmt.Sprintf("127.0.0.1:%v", port), block, 10, 1)
	if err != nil {
		panic(err)
	}

	errorA := errors.New("A")
	err = l.Control(func(conn UT.SuperConn) error {
		fmt.Printf("Listener Control: conn: %v\n", conn)
		return errorA
	})

	if err != errorA {
		t.Fatal(err)
	}

	cli, err := dialEcho(port)
	if err != nil {
		panic(err)
	}

	errorB := errors.New("B")
	err = cli.Control(func(conn UT.PacketConn) error {
		fmt.Printf("Client Control: conn: %v\n", conn)
		return errorB
	})

	if err != errorB {
		t.Fatal(err)
	}
}

func TestSessionReadAfterClosed(t *testing.T) {
	addrS, err := net.ResolveUDPAddr("udp", "127.0.0.1:15678")
	if err != nil {
		panic(err)
	}
	addrC, err := net.ResolveUDPAddr("udp", "127.0.0.1:15679")
	if err != nil {
		panic(err)
	}
	us, _ := U.Dial("udp", addrS, addrC)
	uc, _ := U.Dial("udp", addrC, addrS)
	defer us.Close()
	defer uc.Close()

	knockDoor := func(c net.Conn, myid string) (string, error) {
		c.SetDeadline(time.Now().Add(time.Second * 3))
		_, err := c.Write([]byte(myid))
		c.SetDeadline(time.Time{})
		if err != nil {
			return "", err
		}
		c.SetDeadline(time.Now().Add(time.Second * 3))
		var buf [1024]byte
		n, err := c.Read(buf[:])
		c.SetDeadline(time.Time{})
		return string(buf[:n]), err
	}

	check := func(c1, c2 *UDPSession) {
		done := make(chan struct{}, 1)
		go func() {
			rid, err := knockDoor(c2, "4321")
			done <- struct{}{}
			if err != nil {
				panic(err)
			}
			if rid != "1234" {
				panic("mismatch id")
			}
		}()
		rid, err := knockDoor(c1, "1234")
		if err != nil {
			panic(err)
		}
		if rid != "4321" {
			panic("mismatch id")
		}
		<-done
	}

	c1, err := NewConn3(0, uc.LocalAddr(), nil, 0, 0, us)
	if err != nil {
		panic(err)
	}
	c2, err := NewConn3(0, us.LocalAddr(), nil, 0, 0, uc)
	if err != nil {
		panic(err)
	}
	check(c1, c2)
	c1.Close()
	c2.Close()
	//log.Println("conv id 0 is closed")

	c1, err = NewConn3(4321, uc.LocalAddr(), nil, 0, 0, us)
	if err != nil {
		panic(err)
	}
	c2, err = NewConn3(4321, us.LocalAddr(), nil, 0, 0, uc)
	if err != nil {
		panic(err)
	}
	check(c1, c2)
}
