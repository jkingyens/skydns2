FROM ubuntu:12.04

RUN apt-get update -q
RUN DEBIAN_FRONTEND=noninteractive apt-get install -qy build-essential curl git
RUN curl -s https://go.googlecode.com/files/go1.2.1.src.tar.gz | tar -v -C /usr/local -xz
RUN cd /usr/local/go/src && ./make.bash --no-clean 2>&1
ENV PATH /usr/local/go/bin:$PATH
ENV GOROOT /usr/local/go
ENV GOPATH /work
RUN mkdir -p /work
ADD . /skydns2
RUN cd /skydns2 && go get -d -v ./... && go build 
EXPOSE 53/udp
ENTRYPOINT [ "/skydns2/skydns2", "-etcd=http://172.17.42.1:4001" ]
