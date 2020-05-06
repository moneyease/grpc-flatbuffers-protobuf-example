

generate_fbs:
	flatc --go --grpc fileupload.fbs

generate_proto:
	mkdir fileuploadpb
	protoc fileupload.proto --go_out=plugins=grpc:fileuploadpb

all: clean generate_proto generate_fbs compile_fileupload_client compile_fileupload_server

compile_fileupload_client:
	cd fileupload-client && go build -o ../fileclient  && cd ..

compile_fileupload_server:
	cd fileupload-server && go build -o ../fileserver  && cd ..

clean:
	rm -rf fileserver fileclient fileupload fileuploadpb client-trace.out

.PHONY: clean generate_fbs generate_proto compile compile_fileupload_client compile_fileupload_server
