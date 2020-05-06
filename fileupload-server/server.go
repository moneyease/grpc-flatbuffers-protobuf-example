package main

import (
	fb "../fileupload"
	pb "../fileuploadpb"
	"bytes"
	"crypto/sha256"
	_ "encoding/hex"
	flatbuffers "github.com/google/flatbuffers/go"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"hash"
	_ "io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	_ "os"
	"sync"
	"time"
)

type fbServer struct {
	out   bytes.Buffer
	nPkt  int
	size  int
	fname string
	sum   [32]byte
	begin time.Time
	h     hash.Hash
}
type pbServer fbServer

var tmTaken []time.Duration
var tmUpdate time.Time
var lock sync.RWMutex

var fbAddr = "0.0.0.0:50051"
var pbAddr = "0.0.0.0:50052"

func (s *fbServer) Upload(context context.Context, in *fb.UploadRequest) (*flatbuffers.Builder, error) {
	b := flatbuffers.NewBuilder(0)
	flag := in.Flag()
	tmUpdate = time.Now()
	s.nPkt++
	if flag == fb.FlagFirst {
		s.begin = time.Now()
		s.size = int(in.Size())
		s.fname = string(in.Filename())
		s.h = sha256.New()
		log.Printf("File %s size %d", s.fname, s.size)
	}
	if flag == fb.FlagLast {
		lock.Lock()
		tmTaken = append(tmTaken, time.Since(s.begin))
		lock.Unlock()
		//		log.Printf("%s %s", hex.EncodeToString(s.h.Sum(nil)), os.Args[1])
		s.nPkt = 0
		s.size = 0
	} else {
		file := in.File()
		s.h.Write(file)
	}
	fb.UploadResponseStart(b)
	b.Finish(fb.UploadResponseEnd(b))
	return b, nil
}

func (s *pbServer) Upload(context context.Context, in *pb.UploadRequest) (*pb.UploadResponse, error) {
	flag := in.GetFlag()
	tmUpdate = time.Now()

	if flag == pb.UploadRequest_First {
		s.begin = time.Now()
		s.size = int(in.GetSize())
		s.fname = string(in.GetFilename())
		s.h = sha256.New()
		log.Printf("File %s size %d", s.fname, s.size)
	}
	s.nPkt++
	if flag == pb.UploadRequest_Last {
		lock.Lock()
		tmTaken = append(tmTaken, time.Since(s.begin))
		lock.Unlock()
		//log.Printf("%s %s", hex.EncodeToString(s.h.Sum(nil)), os.Args[1])
		s.nPkt = 0
		s.size = 0
	} else {
		file := in.GetFile()
		s.h.Write(file)
	}
	return &pb.UploadResponse{}, nil
}

func main() {
	done := make(chan bool, 1)
	go func() {
		r := http.NewServeMux()
		// Register pprof handlers
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
		http.ListenAndServe(":9090", r)
	}()

	func() {
		lis, err := net.Listen("tcp", fbAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		ser := grpc.NewServer(grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}))
		fb.RegisterFileUploadServer(ser, &fbServer{})
		go ser.Serve(lis)
	}()

	func() {
		lis, err := net.Listen("tcp", pbAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		ser := grpc.NewServer()
		pb.RegisterFileUploadServer(ser, &pbServer{})
		go ser.Serve(lis)
	}()
	expiry := 5 * time.Second
	go func() {
		tick := time.NewTicker(expiry)
		for {
			select {
			case <-tick.C:
				if time.Since(tmUpdate) > expiry && len(tmTaken) > 0 {
					var min, max, avg time.Duration
					for _, e := range tmTaken {
						if e > max {
							max = e
						}
						if e < min || min == 0 {
							min = e
						}
						avg += e
					}
					log.Printf("Received file %d times in min:%v max:%v avg:%v",
						len(tmTaken), min, max,
						time.Duration(int(avg)/len(tmTaken)))
					tmTaken = nil
				}
			}
		}
	}()
	<-done

}
