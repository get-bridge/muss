FROM alpine

COPY ./entrypoint /entrypoint

CMD ["/entrypoint", "thing"]
