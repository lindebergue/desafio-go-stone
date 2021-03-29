FROM golang:1.16

ENV APP_USER main
ENV APP_ROOT /go/src/main

RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_ROOT && chown -R $APP_USER:$APP_USER $APP_ROOT

USER $APP_USER
WORKDIR $APP_ROOT

COPY . $APP_ROOT
RUN go mod download
RUN go mod verify
RUN go build -o main

ENTRYPOINT [ "./main" ]
