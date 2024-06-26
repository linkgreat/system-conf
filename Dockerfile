FROM ubuntu:20.04 AS runner
RUN apt update
RUN apt install iproute2 net-tools -y


FROM golang:1.21.8-bullseye AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION
RUN echo $VERSION > ./version/ver.txt
RUN cat ./version/ver.txt
#RUN swag init --instanceName iss --pd -d ./ -g ./main.go
RUN go build -a -o systemconf .

FROM runner
WORKDIR /app
COPY --from=builder /src/systemconf ./systemconf

CMD ./systemconf --port=8081