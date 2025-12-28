<h1 align="center">
  PG-REALTIME-CDC
</h1>

<h4 align="center">Aplicação fullstack para captura de dados em tempo real (CDC) com PostgreSQL, Go e React.</h4>

<p align="center">
  <a href="#key-features">Funcionalidades</a> •
  <a href="#api-routes">Rotas da API</a> •
  <a href="#how-to-use">Como Usar</a> •
  <a href="#download">Download</a> •
  <a href="#credits">Créditos</a> •
  <a href="#license">Licença</a>
</p>

## Funcionalidades

* CRUD de mensagens
* Atualização em tempo real via WebSocket
* CDC com PostgreSQL logical replication
* Interface em React
* Backend Go com API REST e WebSocket
* Orquestração via Docker Compose

## Rotas da API

### GET /api/messages
Retorna todas as mensagens.

**Contrato de resposta:**
```json
[
  {
    "id": 1,
    "content": "Olá mundo",
    "created_at": "2025-12-28T01:37:04.832026Z"
  }
]
```

### POST /api/messages
Cria uma nova mensagem.

**Body:**
```json
{
  "content": "Nova mensagem"
}
```

**Contrato de resposta:**
```json
{
  "id": 2,
  "content": "Nova mensagem",
  "created_at": "2025-12-28T01:38:58.468Z"
}
```

### PUT /api/messages/update?id={id}
Atualiza o conteúdo de uma mensagem.

**Body:**
```json
{
  "content": "Mensagem editada"
}
```

**Contrato de resposta:**
```json
{
  "id": 2,
  "content": "Mensagem editada",
  "created_at": "2025-12-28T01:38:58.468Z"
}
```

### DELETE /api/messages/delete?id={id}
Deleta uma mensagem.

**Contrato de resposta:**
```json
{
  "message": "Message deleted successfully"
}
```

### WebSocket /ws
Recebe eventos em tempo real de inserção, atualização e deleção de mensagens.

**Contrato de evento:**
```json
{
  "table": "messages",
  "op": "INSERT|UPDATE|DELETE",
  "data": {
    "id": 2,
    "content": "Mensagem editada"
  },
  "old": {
    "id": 2,
    "content": "Mensagem antiga"
  }
}
```

## Como Usar

Para clonar e rodar esta aplicação, você precisa do [Git](https://git-scm.com), [Docker](https://www.docker.com/) e [Node.js](https://nodejs.org/en/download/) instalados. No terminal:

```bash
# Clone este repositório
$ git clone https://github.com/seu-usuario/pg-realtime-cdc

# Entre no diretório
$ cd pg-realtime-cdc

# Suba o banco e backend
$ docker compose up

# Instale e rode o frontend
$ cd frontend
$ npm install
$ npm run dev
```

Acesse:
- Frontend: [http://localhost:5173](http://localhost:5173)
- Backend REST: [http://localhost:8080/api/messages](http://localhost:8080/api/messages)
- WebSocket: [ws://localhost:8080/ws](ws://localhost:8080/ws)

## Download

Clone o projeto ou baixe via GitHub.

## Créditos

Este software utiliza os seguintes pacotes open source:
- [Go](https://golang.org/)
- [React](https://react.dev/)
- [PostgreSQL](https://www.postgresql.org/)
- [jackc/pglogrepl](https://github.com/jackc/pglogrepl)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [Vite](https://vitejs.dev/)

## Licença

MIT

---

> Projeto por [Seu Nome ou GitHub](https://github.com/seu-usuario)