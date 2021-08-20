package main

import (
	"log"
	"os"

	db "ulambda/debug"
	"ulambda/depproc"
	"ulambda/fsclnt"
	"ulambda/fslibsrv"
	"ulambda/linuxsched"
	"ulambda/memfsd"
)

func main() {
	linuxsched.ScanTopology()
	name := memfsd.MEMFS + "/" + os.Args[1]
	ip, err := fsclnt.LocalIP()
	if err != nil {
		log.Fatalf("%v: no IP %v\n", os.Args[0], err)
	}
	db.Name(name)
	fsd := memfsd.MakeFsd(ip + ":0")
	fsl, err := fslibsrv.InitFs(name, fsd, nil)
	if err != nil {
		log.Fatalf("%v: InitFs failed %v\n", os.Args[0], err)
	}
	sctl := depproc.MakeDepProcCtl(fsl.FsLib, depproc.DEFAULT_JOB_ID)
	sctl.Started(os.Args[1])
	fsd.GetSrv().GetFsServer().GetStats().MakeElastic(fsl.Clnt(), os.Args[1])
	fsd.Serve()
	fsd.GetSrv().GetFsServer().GetStats().Done()
	sctl.Exited(os.Args[1])
	fsl.ExitFs(name)
}
