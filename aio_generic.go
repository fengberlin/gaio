package gaio

import (
	"net"
	"syscall"
)

// event represent a file descriptor event
type event struct {
	ident int  // identifier of this event, usually file descriptor
	r     bool // readable
	w     bool // writable
}

// events from epoll_wait passing to loop,should be in batch for atomicity
type pollerEvents struct {
	events []event
	done   chan struct{}
}

func dupconn(conn net.Conn) (newfd int, err error) {
	sc, ok := conn.(interface {
		SyscallConn() (syscall.RawConn, error)
	})
	if !ok {
		return -1, ErrUnsupported
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return -1, ErrUnsupported
	}
	ec := rc.Control(func(fd uintptr) {
		newfd, err = syscall.Dup(int(fd))
		return
	})

	if ec != nil {
		return -1, ec
	}

	// as we duplicated succesfuly, we're safe to
	// close the original connection
	conn.Close()
	return
}