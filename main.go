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

const VERSION = "1.1"

var cleanup = flag.Bool("cleanup", false, "Clean up converted files")
var output = flag.String("out", "", "Path for converted files to be placed")
var verbose = flag.Bool("verbose", false, "Enable verbose logging")
var force = flag.Bool("force", false, "Overwrite existing files")
var bVersion = flag.Bool("version", false, "Display version")

func init() {
	getopt.Alias("c", "cleanup")
	getopt.Alias("o", "out")
	getopt.Alias("v", "verbose")
	getopt.Alias("f", "force")
	getopt.Alias("V", "version")
	flag.Usage = func() {
		fmt.Printf("Usage:\n%v <options...> [format] [files...]\n\nOptions:\n", os.Args[0])
		getopt.PrintDefaults()
	}
	getopt.Parse()
	vlog("Verbose mode enabled")
}

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
	for _, file := range flag.Args()[1:] {
		convertFile(file, format)
	}
	fmt.Printf("Done! (%vms)\n", time.Since(now).Milliseconds())
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
		println(text)
	}
}
