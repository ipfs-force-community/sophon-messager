FROM filvenus/venus-buildenv AS buildenv

COPY ./go.mod ./venus-messager/go.mod
RUN export GOPROXY=https://goproxy.cn,direct && cd venus-messager   && go mod download 

COPY . ./venus-messager
RUN export GOPROXY=https://goproxy.cn,direct && cd venus-messager && make

RUN cd venus-messager && ldd ./venus-messager


FROM filvenus/venus-runtime



# copy the app from build env
COPY --from=buildenv  /go/venus-messager/venus-messager /app/venus-messager
COPY ./docker/script  /script

EXPOSE 39812

ENTRYPOINT ["/app/venus-messager","run"]
