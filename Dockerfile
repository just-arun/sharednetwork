FROM golang

RUN mkdir /app

ADD . /app

WORKDIR /app

EXPOSE 3030

RUN go build -o main .

CMD ["/app/main"]