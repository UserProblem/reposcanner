FROM golang:1.18 AS build

WORKDIR /go/src/github.com/UserProblem/reposcanner
COPY engine ./engine
COPY go ./go
COPY models ./models
COPY main.go .
COPY go.mod .

ENV CGO_ENABLED=0
RUN go get -d -v ./...

RUN go build -a -installsuffix cgo -o swagger .

FROM alpine AS runtime
COPY --from=build /go/src/github.com/UserProblem/reposcanner/swagger ./
COPY .env ./

EXPOSE 8080/tcp
ENTRYPOINT ["./swagger"]
