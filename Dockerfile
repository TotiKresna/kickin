FROM golang:1.24-alpine

WORKDIR /kickin

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 5000

CMD [ "./main" ]
