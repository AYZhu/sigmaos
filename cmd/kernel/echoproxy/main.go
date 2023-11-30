package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	db "sigmaos/debug"
	"sigmaos/downloaddclnt"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
	"sigmaos/unionrpcclnt"
)

var pipeFile = "/tmp/proxy-in.log"
var pipeOutFile = "/tmp/proxy-out.log"

func main() {
	sc, err := sigmaclnt.NewSigmaClnt(proc.GetProcEnv())
	println(proc.GetProcEnv().KernelID)
	l, err := unionrpcclnt.NewUnionRPCClnt(sc.FsLib, sp.DOWNLOADD, db.ALWAYS, db.ALWAYS).GetSrvs()
	fmt.Println(l)
	ddc, err := downloaddclnt.NewDownloaddClnt(sc.FsLib, proc.GetProcEnv().KernelID)
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
			file.Close()
			out.Close()
			sc.ClntExitOK()
			println("exiting!")
			os.Exit(0)
		}
		if len(text) == 0 {
			continue
		}
		if text[0] == 'x' {
			sc.ClntExitOK()
			println("exiting!")
			out.WriteString("xd\n")
			file.Close()
			out.Close()
			os.Exit(0)
		}
		if text[0] == 'l' {
			print("load at ")
			println(text[1 : len(text)-1])
			ddc.Download(text[1 : len(text)-1])
		}
		out.WriteString("d\n")
	}
}
