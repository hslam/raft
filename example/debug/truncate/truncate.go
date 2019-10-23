package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

// Truncate file to half size.
func truncateFile(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	mem, err := syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}

	size := fi.Size()
	if size <= 1 {
		// Don't need to truncate file if it's too small
		return nil
	}

	trun := size/2 - 1

	//fmt.Printf("size=%d, trun=%d\n", size, trun)
	if trun >= size - 1 {
		trun = size/2
	} else {
		trun = trun + 1
	}

	// Overwrite file content
	copy(mem[0:], mem[trun:])

	err = syscall.Munmap(mem)
	if err != nil {
		return err
	}

	// truncate file
	f.Truncate(fi.Size() - trun)

	// reset file offset
	f.Seek(trun,0)

	return nil
}



func main() {
	dir:="/Users/huangmeng/go/src/default.storage/"
	path := dir+"index"
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Cannot create file")
		log.Fatal(err)
	}
	f.Truncate(1024*1024*1024*32)
	time.Sleep(time.Second*10)
	start:=time.Now().Unix()
	err = truncateFile(f)
	if err != nil {
		fmt.Println("Cannot truncateLog file")
		log.Fatal(err)
	}
	fmt.Println(time.Now().Unix()-start)
	f.Close()
	time.Sleep(time.Minute)
}