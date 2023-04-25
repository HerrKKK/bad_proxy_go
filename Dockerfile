FROM golang:alpine

ADD ./src /bad_proxy_go/
COPY ./conf/server_config.json /bad_proxy_go
WORKDIR /bad_proxy_go
RUN go build -o main \
&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo 'Asia/Shanghai' > /etc/timezone

CMD ["./main", "-c", "proxy_config.json"]
