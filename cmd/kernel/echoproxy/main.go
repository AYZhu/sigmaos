package main

import (
	"bufio"
	"io"
	"os"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
)

var pipeFile = "/tmp/proxy-in.log"
var pipeOutFile = "/tmp/proxy-out.log"

func main() {
	sc, err := sigmaclnt.NewSigmaClnt(proc.GetProcEnv())
	println(proc.GetProcEnv().KernelID)
	println("GO: open read pipe")
	file, err := os.OpenFile(pipeFile, os.O_CREATE|os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		os.Exit(0)
	}
	println("GO: open write pipe")
	out, err := os.OpenFile(pipeOutFile, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0777)
	println("GO: pipes open")
	reader := bufio.NewReader(file)
	if err != nil {
		os.Exit(0)
	}
	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			os.Exit(0)
		}
		if text[0] == 'x' {
			sc.ClntExitOK()
			os.Exit(0)
		}
		if text[0] == 'l' {
			print("load at ")
			println(text)
		}
		out.WriteString("d\n")
	}
}
