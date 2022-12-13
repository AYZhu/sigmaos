package hotel

import (
	"path"

	"sigmaos/cacheclnt"
	db "sigmaos/debug"
	"sigmaos/fslib"
	"sigmaos/proc"
	"sigmaos/procclnt"
)

const (
	HOTEL      = "hotel"
	HOTELDIR   = "name/hotel/"
	MEMFS      = "memfs"
	HTTP_ADDRS = "http-addr"
	NCACHE     = 6
)

func JobDir(job string) string {
	return path.Join(HOTELDIR, job)
}

func JobHTTPAddrsPath(job string) string {
	return path.Join(JobDir(job), HTTP_ADDRS)
}

func MemFsPath(job string) string {
	return path.Join(JobDir(job), MEMFS)
}

func GetJobHTTPAddrs(fsl *fslib.FsLib, job string) ([]string, error) {
	p := JobHTTPAddrsPath(job)
	var addrs []string
	err := fsl.GetFileJson(p, &addrs)
	return addrs, err
}

func InitHotelFs(fsl *fslib.FsLib, jobname string) {
	fsl.MkDir(HOTELDIR, 0777)
	if err := fsl.MkDir(JobDir(jobname), 0777); err != nil {
		db.DFatalf("Mkdir %v err %v\n", JobDir(jobname), err)
	}
}

var HotelSvcs = []string{"user/hotel-userd", "user/hotel-rated",
	"user/hotel-geod", "user/hotel-profd", "user/hotel-searchd",
	"user/hotel-reserved", "user/hotel-recd", "user/hotel-wwwd"}

var ncores = []int{0, 2,
	2, 2, 2,
	4, 0, 2}

func MakeHotelJob(fsl *fslib.FsLib, pclnt *procclnt.ProcClnt, job string, srvs []string, ncache int) (*cacheclnt.CacheClnt, *cacheclnt.CacheMgr, []proc.Tpid, error) {
	var cc *cacheclnt.CacheClnt
	var cm *cacheclnt.CacheMgr
	var err error

	// Init fs.
	InitHotelFs(fsl, job)

	// Create a cache clnt.
	if ncache > 0 {
		cm = cacheclnt.MkCacheMgr(fsl, pclnt, job, ncache)
		cm.StartCache()
		cc, err = cacheclnt.MkCacheClnt(fsl, ncache)
		if err != nil {
			db.DFatalf("Error cacheclnt %v", err)
			return nil, nil, nil, err
		}
	}

	pids := make([]proc.Tpid, 0, len(srvs))

	for i, srv := range srvs {
		p := proc.MakeProc(srv, []string{job})
		p.SetNcore(proc.Tcore(ncores[i]))
		if _, errs := pclnt.SpawnBurst([]*proc.Proc{p}); len(errs) > 0 {
			db.DFatalf("Error burst-spawnn proc %v: %v", p, errs)
			return nil, nil, nil, err
		}
		if err = pclnt.WaitStart(p.Pid); err != nil {
			db.DFatalf("Error spawn proc %v: %v", p, err)
			return nil, nil, nil, err
		}
		pids = append(pids, p.Pid)
	}

	return cc, cm, pids, nil
}
