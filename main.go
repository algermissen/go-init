//go:build (linux && 386) || (darwin && !cgo)

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

func main() {
	fmt.Printf("HALLO 1\n")
	time.Sleep(1 * time.Second)
	cmd := &exec.Cmd{
		Path: "hello",
		Args: []string{"hello"},
	}

	fmt.Printf("Mounting devices...\n")
	must(syscall.Mount("proc", "proc", "proc", 0, ""), "Unable to mount proc")
	must(syscall.Mount("sys", "sys", "sysfs", 0, ""), "Unable to mount sys")
	must(syscall.Mount("dev", "dev", "devtmpfs", 0, ""), "Unable to mount dev")

	must(syscall.Mkdir("/dev/pts", 0600), "Unable to create /dev/pts")
	must(syscall.Mount("dev/pts", "dev/pts", "devpts", 0, ""), "Unable to mount devpts")

	must(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""), "Unable to mount tmpfs")

	fmt.Printf("Devices mounted\n")


	fmt.Printf("Initializing network...\n")
	must(configureEthernet(),"Unable to configure ethernet")
	fmt.Printf("network initialzized\n")

	var err error
	if true {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Millisecond * time.Duration(1000),
				}
				return d.DialContext(ctx, network, "8.8.8.8:53")
			},
		}
		//ip, _ := r.LookupHost(context.Background(), "www.google.com")

		ips, err := r.LookupIP(context.Background(), "ip4", "nytimes.com")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		}
		for _, ip := range ips {
			fmt.Printf(" IN A %s\n", ip.String())
		}
	}

	fmt.Printf("HALLO 1\n")

	err = Run(cmd, os.Stdout)
	if err != nil {
		fmt.Printf("Run: %v", err)
	}

	fmt.Printf("HALLO 2\n")
	time.Sleep(10 * time.Second)
}

func Run(cmd *exec.Cmd, w io.Writer) error {

	if true {
		cmd.Stdin = bytes.NewBuffer(nil)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating StdoutPipe for cmd %v", err)
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("creating StderrPipe for cmd %v", err)
	}

	outScanner := bufio.NewScanner(outPipe)
	go func() {
		for outScanner.Scan() {
			fmt.Fprintf(w, "%s\n", outScanner.Text())
		}
	}()

	errorTextCatched := []string{}
	errScanner := bufio.NewScanner(errPipe)
	go func() {
		for errScanner.Scan() {
			errorTextCatched = append(errorTextCatched, errScanner.Text())
			fmt.Fprintf(w, "ERROR: %s\n", errScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("starting cmd %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("waiting for cmd %v", err)
	}

	return nil
}

func must(err error, msg string) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("%s: %v", msg, err))
}

type socketAddrRequest struct {
	name [unix.IFNAMSIZ]byte
	addr unix.RawSockaddrInet4
}

type socketFlagsRequest struct {
	name  [unix.IFNAMSIZ]byte
	flags uint16
	pad   [22]byte
}

func configureEthernet() error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return errors.Wrap(err, "could not open control socket")
	}

	defer unix.Close(fd)

	// We want to associate an IP address with eth0, then set flags to
	// activate it

	sa := socketAddrRequest{}
	copy(sa.name[:], "eth0")
	sa.addr.Family = unix.AF_INET
	copy(sa.addr.Addr[:], []byte{10, 0, 2, 15})

	// Set address
	if err := ioctl(fd, unix.SIOCSIFADDR, uintptr(unsafe.Pointer(&sa))); err != nil {
		return errors.Wrap(err, "failed setting address for eth0")
	}

	// Set netmask
	copy(sa.addr.Addr[:], []byte{255, 255, 255, 0})
	if err := ioctl(fd, unix.SIOCSIFNETMASK, uintptr(unsafe.Pointer(&sa))); err != nil {
		return errors.Wrap(err, "failed setting netmask for eth0")
	}

	// Get flags
	sf := socketFlagsRequest{}
	sf.name = sa.name
	if err := ioctl(fd, unix.SIOCGIFFLAGS, uintptr(unsafe.Pointer(&sf))); err != nil {
		return errors.Wrap(err, "failed getting flags for eth0")
	}

	sf.flags |= unix.IFF_UP | unix.IFF_RUNNING
	if err := ioctl(fd, unix.SIOCSIFFLAGS, uintptr(unsafe.Pointer(&sf))); err != nil {
		return errors.Wrap(err, "failed getting flags for eth0")
	}

	return nil
}

func ioctl(fd int, code, data uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), code, data)
	if errno != 0 {
		return errno
	}
	return nil
}
