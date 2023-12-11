package main

import (
	"bufio"
	"io"
	"os"
	"path"
	"sigmaos/downloaddclnt"
	"sigmaos/proc"
	"sigmaos/sigmaclnt"
)

var pipeFile = "/tmp/proxy-in.log"
var pipeOutFile = "/tmp/proxy-out.log"

func quitErr(num int, text string, sc *sigmaclnt.SigmaClnt) {
	println("error %v, %v", num, text)
	sc.ClntExit(proc.NewStatusErr("could not initialize Python!", nil))
	os.Exit(0)
}

func main() {
	sc, err := sigmaclnt.NewSigmaClnt(proc.GetProcEnv())
	ddc, err := downloaddclnt.NewDownloaddClnt(sc.FsLib, proc.GetProcEnv().KernelID, proc.GetProcEnv().GetRealmStr())
	file, err := os.OpenFile(pipeFile, os.O_CREATE|os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		os.Exit(0)
	}
	out, err := os.OpenFile(pipeOutFile, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0777)
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
			os.Exit(0)
		}
		if len(text) == 0 {
			continue
		}
		if text[0] == 'x' {
			sc.ClntExitOK()
			out.WriteString("xd\n")
			file.Close()
			out.Close()
			os.Exit(0)
		}
		if text[0] == 'p' {
			if text[1] == 'b' {
				err := ddc.DownloadLib(path.Join("pylib", "Lib"), false)
				if err != nil {
					quitErr(0, err.Error(), sc)
				}
				err = ddc.DownloadLib(path.Join("pylib", "Lib", "encodings"), true)
				if err != nil {
					quitErr(1, err.Error(), sc)
				}
				err = ddc.DownloadLib(path.Join("pylib", "Lib", "importlib"), true)
				if err != nil {
					quitErr(2, err.Error(), sc)
				}
				err = ddc.DownloadLib(path.Join("pylib", "Lib", "splib.py"), false)
				if err != nil {
					quitErr(3, err.Error(), sc)
				}
				err = ddc.DownloadLib(path.Join("pylib", "Lib", "warnings.py"), false)
				if err != nil {
					quitErr(4, err.Error(), sc)
				}
			}
			if text[1] == 'f' {
				err := ddc.DownloadLib(text[2:len(text)-1], false)
				if err != nil {
					out.WriteString("d\n")
				}
			} else if text[1] == 'd' {
				err := ddc.DownloadLib(text[2:len(text)-1], true)
				if err != nil {
					out.WriteString("d\n")
				}
			}
		}
		out.WriteString("d\n")
	}
}
