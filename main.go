package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"rsc.io/getopt"
)

const VERSION = "1.2"

var cleanup = flag.Bool("cleanup", false, "Clean up converted files")
var output = flag.String("out", "", "Path for converted files to be placed")
var verbose = flag.Bool("verbose", false, "Enable verbose logging")
var force = flag.Bool("force", false, "Overwrite existing files")
var bVersion = flag.Bool("version", false, "Display version")
var progress = flag.Bool("progress", false, "Display progress")

func init() {
	getopt.Alias("c", "cleanup")
	getopt.Alias("o", "out")
	getopt.Alias("v", "verbose")
	getopt.Alias("f", "force")
	getopt.Alias("V", "version")
	getopt.Alias("p", "progress")
	flag.Usage = func() {
		fmt.Printf("Usage:\n%v <options...> [format] [files...]\n\nOptions:\n", os.Args[0])
		getopt.PrintDefaults()
	}
	getopt.Parse()
	if *verbose {
		println("Verbose mode enabled")
	}
}

var i int
var leng int = 1

func main() {
	if *bVersion {
		fmt.Printf("fconvert: %v\n", VERSION)
	}
	if flag.NArg() < 2 {
		fmt.Println("Error: not enough arguments")
		fmt.Println()
		flag.Usage()
		os.Exit(2)
	}

	now := time.Now()
	format := flag.Args()[0]
	leng = len(flag.Args()[1:])
	if *progress {
		fmt.Printf("(0.00%%) 0/%v\n", leng)
	}
	for _i, file := range flag.Args()[1:] {
		i = _i + 1
		convertFile(file, format)
		if *progress {
			fmt.Print("\033[A")
			fmt.Print("\033[G")
			fmt.Print("\033[2K")
			fmt.Printf("(%.2f%%) %v/%v\n", float32(i)/float32(leng)*100, i, leng)
		}
	}
	fmt.Printf("Done! (%v)\n", time.Since(now))
}

func convertFile(fileName string, format string) {
	vlog("Converting " + fileName + "...")
	if *output != "" {
		if _, err := os.Stat(*output); err != nil {
			vlog("	Making path: " + *output)
			os.MkdirAll(*output, 0755)
		}
	}

	newName := path.Join(*output, formatName(fileName, format))
	if _, err := os.Lstat(newName); err == nil {
		if *force {
			vlog("	Deleting " + newName + "...")
			if err := os.Remove(newName); err != nil {
				println("Can't delete: " + err.Error())
				return
			}
		} else {
			vlog("	Skipping " + fileName)
			return
		}
	}

	cmd := generateCommand(fileName, format)
	errb := new(bytes.Buffer)
	cmd.Stderr = errb
	err := cmd.Run()
	if err != nil {
		str := errb.String()
		fmt.Printf("ERROR on %v\n", fileName)
		if strings.Contains(str, "already exists") {
			fmt.Printf("Failed: %v already exists\n", newName)
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
			fmt.Printf("	Unable to remove %v (%v)\n", fileName, err.Error())
		}
		vlog("	Deleted " + fileName)
	}

	vlog("	Done ceonverting " + fileName + " -> " + newName)
}

func generateCommand(from, fmt string) *exec.Cmd {
	to := path.Join(*output, path.Base(formatName(from, fmt)))
	var cmd *exec.Cmd

	switch fmt {
	case "jxl":
		vlog("	Using cjxl")
		cmd = exec.Command("cjxl", from, to)
	case "mp4":
		vlog("	Using ffmpeg (hevc)")
		cmd = exec.Command("ffmpeg", "-hide_banner", "-hwaccel", "auto", "-i", from, "-c:v", "hevc", to)
	default:
		vlog("	Using ffmpeg")
		cmd = exec.Command("ffmpeg", "-hide_banner", "-hwaccel", "auto", "-i", from, to)
	}

	return cmd
}

func formatName(fileName string, format string) string {
	split := strings.Split(fileName, ".")
	name := strings.Join(split[:len(split)-1], ".")
	return (name + "." + format)
}

func vlog(text string) {
	if *verbose {
		if *progress {
			fmt.Print("\033[A")
			fmt.Print("\033[G")
			fmt.Print("\033[2K")
			fmt.Println(text)
			fmt.Printf("(%.2f%%) %v/%v\n", float32(i)/float32(leng)*100, i, leng)
		} else {
			fmt.Println(text)
		}
	}
}
