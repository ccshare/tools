FROM scratch

ADD . /
EXPOSE  80/tcp
WORKDIR /
VOLUME  /tmp
ENV ZONEINFO=/zoneinfo.zip \
	TZ=Asia/Shanghai

CMD ["/authd"]
