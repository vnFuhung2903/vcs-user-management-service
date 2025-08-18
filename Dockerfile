FROM golang:1.24.2

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8085

CMD [ "go", "run", "./cmd/main.go" ]