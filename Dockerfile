
FROM golang:1.10.0
RUN go get github.com/codegangsta/negroni \
           github.com/gorilla/mux \
           github.com/xyproto/simpleredis \
	   github.com/go-sql-driver/mysql

WORKDIR /app
ADD ./main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM scratch
WORKDIR /app
COPY --from=0 /app/main .
CMD ["/app/main"]
EXPOSE 3000
