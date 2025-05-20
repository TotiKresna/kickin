FROM golang:1.24-alpine

WORKDIR /kickin

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p assets/image

RUN go build -o main .

EXPOSE 5000

CMD [ "./main" ]
