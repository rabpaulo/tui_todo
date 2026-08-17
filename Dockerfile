# Stage 1: Build the Go binary
FROM golang:alpine AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos necessários
COPY go.mod .
COPY main.go .

# Baixa as dependências e arruma o go.mod/go.sum
RUN go mod tidy

# Compila o executável (binário estático, otimizado)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o tui-todo main.go

# Stage 2: Imagem final minimalista
FROM alpine:latest

WORKDIR /root/

# Copia o binário pronto da imagem anterior
COPY --from=builder /app/tui-todo .

# Configura as variáveis de ambiente necessárias para que o terminal exiba as cores corretamente
ENV TERM=xterm-256color

# Comando para iniciar a aplicação
CMD ["./tui-todo"]
