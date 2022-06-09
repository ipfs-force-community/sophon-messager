FROM filvenus/venus-buildenv AS buildenv

RUN git clone https://github.com/filecoin-project/venus-messager.git --depth 1 
RUN export GOPROXY=https://goproxy.cn && cd venus-messager && make

RUN cd venus-messager && ldd ./venus-messager


FROM filvenus/venus-runtime

# DIR for app
WORKDIR /app

# copy the app from build env
COPY --from=buildenv  /go/venus-messager/venus-messager /app/venus-messager


# 拷贝依赖库
COPY --from=buildenv  /usr/lib/x86_64-linux-gnu/libhwloc.so.5 \
/usr/lib/x86_64-linux-gnu/libOpenCL.so.1 \
/lib/x86_64-linux-gnu/libgcc_s.so.1 \
/lib/x86_64-linux-gnu/libutil.so.1 \
/lib/x86_64-linux-gnu/librt.so.1 \
/lib/x86_64-linux-gnu/libpthread.so.0 \
/lib/x86_64-linux-gnu/libm.so.6 \
/lib/x86_64-linux-gnu/libdl.so.2 \
/lib/x86_64-linux-gnu/libc.so.6 \
/usr/lib/x86_64-linux-gnu/libnuma.so.1 \
/usr/lib/x86_64-linux-gnu/libltdl.so.7 \
                        /lib/

COPY ./docker/script  /script

EXPOSE 39812

ENTRYPOINT ["/app/venus-messager","run"]
