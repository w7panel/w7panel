
`
crane pull docker.wos.w7.com/public/proxy:go-2.0.5 ./proxy-2.0.5
or 
skopeo copy docker://alpine:latest oci:./proxy

crane push --insecure ./proxy 127.0.0.1:5000/test2:latest --verbose
or
crane push --insecure ./proxy2.0.5 127.0.0.1:5000/test2:latest --verbose
`