FROM alpine:3.7

COPY ./entrypoint.sh /
COPY ./bin /usr/local/bin
RUN chmod +x entrypoint.sh
RUN	CGO_ENABLED=0 GOOS=linux go build cmd/oc-historical_logs.go
RUN chmod +x oc-historical_logs
COPY ./oc-historical_logs /usr/local/bin
ENTRYPOINT ["/entrypoint.sh"]
CMD ["log-exploration-oc-plugin"]