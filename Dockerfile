FROM alpine:latest
MAINTAINER jiangpengfei <jiangpengfei@xinhuotech.com>

COPY ./build/lazykube /lazykube

ENTRYPOINT [ "./lazykube" ]

EXPOSE 443
