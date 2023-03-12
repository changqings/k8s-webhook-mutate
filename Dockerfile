FROM  ccr.ccs.tencentyun.com/lzwk/ubuntu:v20.04-https


WORKDIR /app

# use volumeMounts
#COPY config config


ADD k8s-webhook-test /usr/local/bin/

