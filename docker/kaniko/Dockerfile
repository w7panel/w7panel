# FROM ccr.ccs.tencentyun.com/afan-public/kaniko:debug
FROM registry.cn-shanghai.aliyuncs.com/w7/kaniko:debug

ADD ./curl-amd64 /kaniko/curl
ADD ./start.sh /kaniko
ADD ./buildimage.sh /kaniko
RUN chmod +x /kaniko/curl && chmod +x /kaniko/start.sh && chmod +x /kaniko/buildimage.sh

#docker build -t ccr.ccs.tencentyun.com/we7team/tty:buildimage .
ENTRYPOINT ["/bin/sh"]

#docker run -it --env-file /home/afan/buildenv  ccr.ccs.tencentyun.com/we7team/tty:buildimage /kaniko/buildimage.sh
