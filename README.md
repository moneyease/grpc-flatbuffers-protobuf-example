
# gRPC Flatbuffers and Protocol Buffers Example

A simple bookmarking service defined in the FlatBuffers and Protocol Buffers IDL, and creation of gRPC server interfaces and client stubs. (`fileupload.fbs` `fileupload.proto`)

## Instructions

### Compile 
```
make all
```

#### Start Server
Server is listening on 2 ports one each for Flatbuffers and Protocol Buffers handles
```
./fileserver
```

#### Client uploading files
```
./client -f <filename> -m proto|flat -c <num iteration>
```

## Resources

https://google.github.io/flatbuffers/flatbuffers_guide_use_go.html
https://developers.google.com/protocol-buffers/docs/gotutorial
