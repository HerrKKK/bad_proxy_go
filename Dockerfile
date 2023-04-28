FROM golang:alpine

ADD ./src /bad_proxy_go/
COPY ./conf/server_config.json /bad_proxy_go/proxy_config.json
COPY ./conf/rules.dat /bad_proxy_go/rules.dat
WORKDIR /bad_proxy_go
RUN go build -o bad_proxy \
&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo 'Asia/Shanghai' > /etc/timezone

CMD ["./bad_proxy", "run", "--config", "proxy_config.json", "--router-path", "rules.dat"]
