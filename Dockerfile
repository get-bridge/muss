FROM golang:1.16-alpine

RUN apk --no-cache add \
    bash \
    gcc \
    musl-dev \
    git \
    make \
  && addgroup -S docker \
  && adduser -S docker -G docker \
  && mkdir /src \
  && chown docker:docker /src

USER docker

WORKDIR /src
COPY --chown=docker:docker . ./
RUN mkdir build coverage \
# Don't let any unexpected writes happen.
  && chmod a-w -R . \
# ... but we need to write in these folders.
  && chmod u+w build coverage \
# Install build and test deps.
  && make build && go test -i
