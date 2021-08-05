FROM golang:1.14-alpine
ENV GOPROXY=https://goproxy.cn
#ENV token=""
WORKDIR /build
COPY . .
EXPOSE 8088
RUN mkdir /app 
RUN  go mod tidy 
RUN go build -o /app/alertGo alertGo.go
WORKDIR /app
CMD ["/app/alertGo"]

