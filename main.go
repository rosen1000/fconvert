package main

// #include <./lib.h>
// typedef int (*intFunc) ();
//
// int
// bridge_int_func(intFunc f)
// {
//		return f();
// }
//
// int fortytwo()
// {
//	    return 42;
// }
import "C"

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"rsc.io/getopt"
	"strings"
	"sync"
	"time"
)

var cleanup = flag.Bool("cleanup", false, "Clean up converted files")
var wg = new(sync.WaitGroup)

func init() {
	// f := C.intFunc(C.fortytwo)
	// fmt.Println(C.bridge_int_func(f))
	fmt.Println(C.two())
	getopt.Alias("c", "cleanup")
	flag.Usage = func() {
		fmt.Printf("Usage:\n%v <options...> [format] [files...]\n\nOptions:\n", os.Args[0])
		getopt.PrintDefaults()
	}
	getopt.Parse()
}

func main() {
	fmt.Println(*cleanup)
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
	// TODO: Chain command benchmark?
	// ffmpeg -i file1.jpg file1.png -i file2.jpg file2.png
	cmd := exec.Command("ffmpeg", "-hide_banner", "-i", fileName, formatName(fileName, format))
	err := cmd.Run()
	if err != nil {
		output, _ := cmd.Output()
		fmt.Println(output)
		fmt.Println(err.Error())
		os.Exit(1)
	}
	wg.Done()

	// fmt.Printf("Formated %v\n", fileName)
}

func formatName(fileName string, format string) string {
	split := strings.Split(fileName, ".")
	name := strings.Join(split[:len(split)-1], ".")
	return (name + "." + format)
}
