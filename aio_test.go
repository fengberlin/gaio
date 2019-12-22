package gaio

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func init() {

	go http.ListenAndServe(":6060", nil)
}

func echoServer(t testing.TB) net.Listener {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	w, err := CreateWatcher()
	if err != nil {
		t.Fatal(err)
	}

	rx := make([]byte, 128)

	chRx := make(chan OpResult)
	chTx := make(chan OpResult)

	go func() {
		for {
			select {
			case res := <-chRx:
				if res.Size > 0 {
					tx := make([]byte, res.Size)
					copy(tx, rx[:res.Size])
					w.Write(res.Fd, tx[:res.Size], chTx)
				}
			case res := <-chTx:
				w.Read(res.Fd, rx, chRx)
			}
		}
	}()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				return
			}

			fd, err := w.Watch(conn)
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("watching", conn.RemoteAddr(), "fd:", fd)

			err = w.Read(fd, rx, chRx)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()
	return ln
}

func TestEcho(t *testing.T) {
	ln := echoServer(t)
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("hello world")
	rx := make([]byte, len(tx))

	conn.Write(tx)
	t.Log("tx:", string(tx))
	_, err = conn.Read(rx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("rx:", string(tx))
}

func BenchmarkEcho(b *testing.B) {
	ln := echoServer(b)

	addr, _ := net.ResolveTCPAddr("tcp", ln.Addr().String())
	tx := []byte("hello world")
	rx := make([]byte, len(tx))

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		b.Fatal(err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		conn.Write(tx)
		conn.Read(rx)
		//		log.Println(i, b.N)
	}
}