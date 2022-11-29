package dbd

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"sigmaos/dbd/proto"
	"sigmaos/debug"
)

type Server struct {
	db   *sql.DB
	rows *sql.Rows
}

func mkServer(dbdaddr string) (*Server, error) {
	s := &Server{}
	db, error := sql.Open("mysql", "sigma:sigmaos@tcp("+dbdaddr+")/sigmaos")
	if error != nil {
		return nil, error
	}
	s.db = db
	error = s.db.Ping()
	if error != nil {
		debug.DFatalf("db.Ping err %v\n", error)
	}
	return s, nil
}

func (s *Server) doQuery(arg string, rep *[]byte) error {
	debug.DPrintf("DBSRV", "doQuery: %v\n", arg)
	rb, err := doQuery(s.db, arg)
	if err != nil {
		return err
	}
	*rep = rb
	return nil
}

func (s *Server) Query(req proto.DBRequest, rep *proto.DBResult) error {
	err := s.doQuery(req.Cmd, &rep.Res)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Exec(req proto.DBRequest, rep *proto.DBResult) error {
	err := s.doQuery(req.Cmd, &rep.Res)
	if err != nil {
		return err
	}
	return nil
}
