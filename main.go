package main

import (
	"fmt"
	"context"
	"time"
	"bufio"
	"os"
	"io"
	"net"
	"bytes"
	"os/exec"
	"syscall"
)

func must(err error, msg string) {
	if err == nil {
		return
	}
	panic(fmt.Sprintf("%s: %v",msg,err))
}

func main() {
	fmt.Printf("HALLO 1\n")
	time.Sleep(1 * time.Second)
	cmd := &exec.Cmd{
		Path: "hello",
		Args: []string{"hello"},
	}

	must(syscall.Mount("proc", "proc", "proc", 0, ""), "Unable to mount proc")
	must(syscall.Mount("sys", "sys", "sysfs", 0, ""), "Unable to mount sys")
	must(syscall.Mount("dev", "dev", "devtmpfs", 0, ""), "Unable to mount dev")

	must(syscall.Mkdir("/dev/pts",0600),"Unable to create /dev/pts")
	must(syscall.Mount("dev/pts", "dev/pts", "devpts", 0, ""), "Unable to mount devpts")

	must(syscall.Mount("tmpfs", "/tmp", "tmpfs", 0, ""), "Unable to mount tmpfs")


	/*
       - initialize network card
       - set ip routing
       - determine or set host IP (DHCP or static)

       os.Exec("ifup","eth0"....)
       */


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



	ips, err := r.LookupIP(context.Background(),"ip4","nytimes.com")
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
