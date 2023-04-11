
FROM golang:1.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /mobile-farm-chat

EXPOSE 8080

CMD [ "/mobile-farm-chat" ]
