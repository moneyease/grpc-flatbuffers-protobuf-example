package main

import (
	fb "../fileupload"
	pb "../fileuploadpb"
	"context"
	"flag"
	flatbuffers "github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime/trace"
	"time"
)

var fbAddr = "0.0.0.0:50051"
var pbAddr = "0.0.0.0:50052"

func fbSendFile(fname string) {
	conn, err := grpc.Dial(fbAddr, grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// pack flatbuffer
	b := flatbuffers.NewBuilder(0)

	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("--> failed to open file")
		return
	}
	fInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("--> failed to stat file")
		return
	}

	ctx := context.Background()
	client := fb.NewFileUploadClient(conn)

	buf := make([]byte, 16*1024)
	var start bool
	var last bool
	var nPkts int
	for {
		begin := time.Now()
		n, err := file.Read(buf)
		b.Reset()
		nPkts++
		if n > 0 {
			file_pos := b.CreateByteString(buf)
			name_pos := b.CreateString(fname)

			fb.UploadRequestStart(b)
			fb.UploadRequestAddFile(b, file_pos)
			if start == false {
				fb.UploadRequestAddFilename(b, name_pos)
				fb.UploadRequestAddFlag(b, fb.FlagFirst)
				fb.UploadRequestAddSize(b, fInfo.Size())
				start = true
			}
			b.Finish(fb.UploadRequestEnd(b))
			//b.Bytes[b.Head():]
		}

		if err == io.EOF {
			fb.UploadRequestStart(b)
			fb.UploadRequestAddFlag(b, fb.FlagLast)
			b.Finish(fb.UploadRequestEnd(b))
			last = true
		}
		// send over grpc
		_, err = client.Upload(ctx, b)
		if err != nil {
			log.Fatalf("Retrieve client failed: %v", err)
		}
		if last {
			tmTaken = append(tmTaken, time.Since(begin))
			nPkts = 0
			break
		}
	}
}

func setSize(r *pb.UploadRequest, size uint32) {
	r.Size = size
}

func setFlag(r *pb.UploadRequest, flag pb.UploadRequest_Flag) {
	r.Flag = flag
}

func setFilename(r *pb.UploadRequest, name string) {
	r.Filename = name
}

func setFile(r *pb.UploadRequest, buf []byte) {
	r.File = buf
}

func pbSendFile(fname string) {
	conn, err := grpc.Dial(pbAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("--> failed to open file")
		return
	}

	fInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("--> failed to stat file")
		return
	}

	ctx := context.Background()
	client := pb.NewFileUploadClient(conn)

	buf := make([]byte, 16*1024)
	var start bool
	var last bool
	var nPkts int
	for {
		begin := time.Now()
		n, err := file.Read(buf)
		r := &pb.UploadRequest{}
		nPkts++
		if n > 0 {
			setFile(r, buf)
			if start == false {
				setFilename(r, fname)
				setFlag(r, pb.UploadRequest_First)
				setSize(r, uint32(fInfo.Size()))
				start = true
			}
			//b.Bytes[b.Head():]
		}

		if err == io.EOF {
			setFlag(r, pb.UploadRequest_Last)
			last = true
		}

		_, err = client.Upload(ctx, r)
		if err != nil {
			log.Fatalf("Fail to get the grpc stream, Error:%v", err)
		}
		if last {
			tmTaken = append(tmTaken, time.Since(begin))
			nPkts = 0
			break
		}
	}
}

var tmTaken []time.Duration
var (
	mode  *string
	file  *string
	count *int
)

func init() {
	count = flag.Int("c", 1, "num test run")
	mode = flag.String("m", "", "set mode to proto/flat")
	file = flag.String("f", "", "specify a file")
}
func main() {
	flag.Parse()
	done := make(chan bool, 1)
	if *file == "" {
		log.Println("Specify a file")
		return
	}
	if *mode == "" {
		log.Println("Specify a mode 'proto' or 'flat'")
		return
	}
	f, err := os.Create("client-trace.out")
	if err != nil {
		panic(err)
	}
	go func() {
		r := http.NewServeMux()
		// Register pprof handlers
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
		http.ListenAndServe(":9091", r)
	}()
	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	if *mode == "flat" {
		for i := 0; i < *count; i++ {
			fbSendFile(*file)
		}
	} else if *mode == "proto" {
		for i := 0; i < *count; i++ {
			pbSendFile(*file)
		}
	} else {
		log.Printf("Invalid option %s - specify a mode 'proto' or 'flat'\n", *mode)
		return
	}
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
	log.Printf("Sent file %d time in min:%v max:%v avg:%v", count, min, max, time.Duration(int(avg)/len(tmTaken)))

	trace.Stop()
	f.Close()
	log.Printf("Done.. (Ctl-C to exit)")
	<-done
}
