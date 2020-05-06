FROM alpine:3.9.2
#RUN apk add bash tzdata python2 python2-dev libcurl py2-pip curl curl-dev openssl-dev gcc libc-dev
#COPY deployment/requirements.txt /opt/app/requirements.txt
RUN mkdir -p /app
WORKDIR /app
EXPOSE 50051
EXPOSE 50052
#RUN pip install -r requirements.txt
ENV TZ America/Los_Angeles
# adding applications in deployment into /usr/local/bin
ADD fileserver /app
COPY fileserver /app
RUN chown nobody:nogroup /app
USER nobody
ENTRYPOINT /app/fileserver
#ENTRYPOINT sleep 1000
