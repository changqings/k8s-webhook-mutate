# k8s-webhook-test

```
use cert-manager generate tls
生成ca时要使用一个ca-tls
openssl genrsa -out ca.key 2048
openssl req -x509 --newkey rsa:2048 -new -nodes -key ca.key -days 3650 -reqexts v3_req -extensions v3_ca -out ca.crt
kubectl create secert tls ca-key-pair --cert=ca.crt --key=ca.key --namespace=default
kubectl apply -f install-on-k8s/issuer.yaml
```

调试过程总结:
- 了解webhook的执行流程及所处的环节
- 了解net/http包的使用,了解handler{w,r}的处理逻辑，接收请求数据，处理后，写入response
- 了解处理的数据的结构admissionReview{}及数据的必备字段，这里只需要返回admissionReview.response即可，不需要追加原ar.Request
- 先找一个可以在线上运行的demo，参考调试，不要过多的参考项目，尤其是哪些自己没有验证过的
- 适当打印日志及数据结构的实体
- json.Marshal和Unmarshal的使用,Unmarshal解析到指针，marshal使用实体
