package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"rsc.io/getopt"
)

var cleanup = flag.Bool("cleanup", false, "Clean up converted files")
var wg = new(sync.WaitGroup)

func init() {
	getopt.Alias("c", "cleanup")
	flag.Usage = func() {
		fmt.Printf("Usage:\n%v <options...> [format] [files...]\n\nOptions:\n", os.Args[0])
		getopt.PrintDefaults()
	}
	getopt.Parse()
}

func main() {
	now := time.Now()
	format := flag.Args()[0]
	for _, file := range flag.Args()[1:] {
		wg.Add(1)
		go convertFile(file, format)
	}
	wg.Wait()
	fmt.Printf("Done! (%vms)\n", time.Now().Sub(now).Milliseconds())
}

func convertFile(fileName string, format string) {
	defer wg.Done()
	cmd := exec.Command("ffmpeg", "-hide_banner", "-i", fileName, formatName(fileName, format))
	errb := new(bytes.Buffer)
	cmd.Stderr = errb
	err := cmd.Run()
	if err != nil {
		str := errb.String()
		if strings.Contains(str, "already exists") {
			fmt.Printf("Failed: %v already exists\n", formatName(fileName, format))
		} else if strings.Contains(str, "No such") {
			fmt.Printf("Failed: %v doesn't exist\n", fileName)
		} else {
			fmt.Println(errb.String())
		}
		return
	}

	if *cleanup {
		err := os.Remove(fileName)
		if err != nil {
			fmt.Printf("Unable to remove %v (%v)\n", fileName, err.Error())
		}
	}
}

func formatName(fileName string, format string) string {
	split := strings.Split(fileName, ".")
	name := strings.Join(split[:len(split)-1], ".")
	return (name + "." + format)
}
