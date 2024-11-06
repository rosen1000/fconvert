package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"rsc.io/getopt"
)

var cleanup = flag.Bool("cleanup", false, "Clean up converted files")
var output = flag.String("out", "", "Path for converted files to be placed")
var verbose = flag.Bool("verbose", false, "Enable verbose logging")
var wg = new(sync.WaitGroup)

func init() {
	getopt.Alias("c", "cleanup")
	getopt.Alias("o", "out")
	getopt.Alias("v", "verbose")
	flag.Usage = func() {
		fmt.Printf("Usage:\n%v <options...> [format] [files...]\n\nOptions:\n", os.Args[0])
		getopt.PrintDefaults()
	}
	getopt.Parse()
	vlog("Verbose mode enabled")
}

func main() {
	if flag.NArg() < 2 {
		fmt.Println("Error: not enough arguments")
		fmt.Println()
		flag.Usage()
		os.Exit(2)
	}

	now := time.Now()
	format := flag.Args()[0]
	for _, file := range flag.Args()[1:] {
		wg.Add(1)
		go convertFile(file, format)
	}
	wg.Wait()
	fmt.Printf("Done! (%vms)\n", time.Since(now).Milliseconds())
}

func convertFile(fileName string, format string) {
	vlog("Converting " + fileName + "...")
	defer wg.Done()
	if *output != "" {
		if _, err := os.Stat(*output); err != nil {
			vlog("Making path: " + *output)
			os.MkdirAll(*output, 0755)
		}
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-i", fileName, path.Join(*output, path.Base(formatName(fileName, format))))
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
		vlog("Deleting" + fileName + "...")
		err := os.Remove(fileName)
		if err != nil {
			fmt.Printf("Unable to remove %v (%v)\n", fileName, err.Error())
		}
		vlog("Done deleting " + fileName)
	}

	vlog("Done ceonverting " + fileName + " -> " + formatName(fileName, format))
}

func formatName(fileName string, format string) string {
	split := strings.Split(fileName, ".")
	name := strings.Join(split[:len(split)-1], ".")
	return (name + "." + format)
}

func vlog(text string) {
	if *verbose {
		println(text)
	}
}
