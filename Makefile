

generate_fbs:
	flatc --go --grpc fileupload.fbs

generate_proto:
	mkdir fileuploadpb
	protoc fileupload.proto --go_out=plugins=grpc:fileuploadpb

all: clean generate_proto generate_fbs compile_fileupload_client compile_fileupload_server

deploy : export GOOS=linux
deploy : compile_fileupload_server
	@echo "Docker clean previous instance.."
	@docker stop fileserver || true && docker rm fileserver || true
	@echo "Docker remove existing images.."
	@docker rmi docker.fileserver -f
	@echo "Docker building image.."
	@docker build --force-rm=true -t docker.fileserver . -f Dockerfile
	@echo "Docker image is running.."
	@docker run -it -p 50051:50051 -p 50052:50052 -p 9090:9090 --name=fileserver  docker.fileserver:latest

compile_fileupload_client:
	cd fileupload-client && go build -o ../fileclient  && cd ..

compile_fileupload_server:
	cd fileupload-server && go build -o ../fileserver  && cd ..

clean:
	rm -rf fileserver fileclient fileupload fileuploadpb client-trace.out

.PHONY: clean generate_fbs generate_proto compile compile_fileupload_client compile_fileupload_server
