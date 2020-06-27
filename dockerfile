FROM go 1.4
RUN mkdir /LetsConnect
ADD ./LetsConnect
WORKDIR /LetsConnect

RUN go mod download
RUN go build -o main ./...

CMD ["/LetsConnect/main"]