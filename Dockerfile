FROM golang:stretch as build

ADD . /go/node_exporter_hostname
WORKDIR /go/node_exporter_hostname

RUN go get
RUN go build -ldflags="-s -w"
RUN chmod +x node_exporter_hostname

FROM prom/node-exporter:latest

COPY --from=build /go/node_exporter_hostname/node_exporter_hostname /bin/node_exporter_hostname

ENTRYPOINT [ "/bin/node_exporter_hostname" ]
