FROM  debian:12.5

WORKDIR /app

# use volumeMounts
#COPY config config


ADD k8s-webhook-mutate /usr/local/bin/

