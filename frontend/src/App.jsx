import { useEffect, useState } from "react";
import "./App.css";

const API_URL = "http://localhost:8080/api/messages";
const WS_URL = "ws://localhost:8080/ws";

export default function App() {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [wsConnected, setWsConnected] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [editContent, setEditContent] = useState("");
  const [newContent, setNewContent] = useState("");

  useEffect(() => {
    fetchMessages();
  }, []);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);

    ws.onopen = () => {
      console.log("WebSocket connected");
      setWsConnected(true);
    };

    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data);
        handleRealtimeEvent(event);
      } catch (err) {
        console.error("Error parsing WebSocket message:", err);
      }
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
      setWsConnected(false);
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
      setWsConnected(false);
    };

    return () => {
      ws.close();
    };
  }, []);

  const fetchMessages = async () => {
    try {
      setLoading(true);
      const response = await fetch(API_URL);
      if (!response.ok) throw new Error("Failed to fetch messages");
      const data = await response.json();
      setMessages(data);
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error("Error fetching messages:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleRealtimeEvent = (event) => {
    if (event.table !== "messages") return;

    switch (event.op) {
      case "INSERT":
        if (event.data && event.data.id) {
          setMessages((prev = []) => [
            {
              id: parseInt(event.data.id),
              content: event.data.content || "",
              created_at: new Date().toISOString(),
            },
            ...prev?.filter(msg => msg.id !== parseInt(event.data.id)),
          ]);
        }
        break;
      case "UPDATE":
        if (event.data && event.data.id) {
          setMessages((prev) =>
            prev.map((msg) =>
              msg.id === parseInt(event.data.id)
                ? { ...msg, content: event.data.content || msg.content }
                : msg
            )
          );
        }
        break;
      case "DELETE":
        if (event.data && event.data.id) {
          setMessages((prev) =>
            prev?.filter((msg) => msg.id !== parseInt(event.data.id))
          );
        }
        break;
      default:
        fetchMessages();
    }
  };

  const createMessage = async (e) => {
    e.preventDefault();
    if (!newContent.trim()) return;

    try {
      const response = await fetch(API_URL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: newContent.trim() }),
      });

      if (!response.ok) throw new Error("Failed to create message");

      const newMessage = await response.json();
      setNewContent("");
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error("Error creating message:", err);
    }
  };

  const startEdit = (message) => {
    setEditingId(message.id);
    setEditContent(message.content);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditContent("");
  };

  const updateMessage = async (id) => {
    if (!editContent.trim()) return;

    try {
      const response = await fetch(`${API_URL}/update?id=${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content: editContent.trim() }),
      });

      if (!response.ok) throw new Error("Failed to update message");

      const updated = await response.json();
      setMessages((prev) =>
        prev.map((msg) => (msg.id === id ? updated : msg))
      );
      setEditingId(null);
      setEditContent("");
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error("Error updating message:", err);
    }
  };

  const deleteMessage = async (id) => {
    if (!confirm("Tem certeza que deseja deletar esta mensagem?")) return;

    try {
      const response = await fetch(`${API_URL}/delete?id=${id}`, {
        method: "DELETE",
      });

      if (!response.ok) throw new Error("Failed to delete message");

      setMessages((prev) => prev.filter((msg) => msg.id !== id));
      setError(null);
    } catch (err) {
      setError(err.message);
      console.error("Error deleting message:", err);
    }
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat("pt-BR", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  };

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-content">
          <h1>
            <span className="icon">⚡</span>
            Real-time CDC Messages
          </h1>
          <div className={`ws-status ${wsConnected ? "connected" : "disconnected"}`}>
            <span className="status-dot"></span>
            {wsConnected ? "Conectado" : "Desconectado"}
          </div>
        </div>
      </header>

      <main className="app-main">
        <div className="container">
          <section className="card create-card">
            <h2>Criar Nova Mensagem</h2>
            <form onSubmit={createMessage} className="message-form">
              <textarea
                value={newContent}
                onChange={(e) => setNewContent(e.target.value)}
                placeholder="Digite sua mensagem aqui..."
                rows="3"
                className="message-input"
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    if (newContent.trim()) {
                      createMessage(e);
                    }
                  }
                }}
              />
              <button type="submit" className="btn btn-primary" disabled={!newContent.trim()}>
                <span className="btn-icon">➕</span>
                Criar Mensagem
              </button>
            </form>
          </section>

          {error && (
            <div className="alert alert-error">
              <span className="alert-icon">⚠️</span>
              {error}
            </div>
          )}

          <section className="card messages-card">
            <div className="messages-header">
              <h2>Mensagens ({messages?.length ?? 0})</h2>
              <button onClick={fetchMessages} className="btn btn-secondary" disabled={loading}>
                {loading ? "Atualizando..." : "🔄 Atualizar"}
              </button>
            </div>

            {loading && messages?.length === 0 ? (
              <div className="loading">Carregando mensagens...</div>
            ) : messages?.length === 0 ? (
              <div className="empty-state">
                <span className="empty-icon">📭</span>
                <p>Nenhuma mensagem ainda. Crie uma nova mensagem acima!</p>
              </div>
            ) : (
              <div className="messages-list">
                {messages?.map((message) => (
                  <div key={message.id} className="message-item">
                    {editingId === message.id ? (
                      <div className="message-edit">
                        <textarea
                          value={editContent}
                          onChange={(e) => setEditContent(e.target.value)}
                          className="message-input"
                          rows="2"
                          autoFocus
                        />
                        <div className="message-actions">
                          <button
                            onClick={() => updateMessage(message.id)}
                            className="btn btn-success btn-sm"
                            disabled={!editContent.trim()}
                          >
                            💾 Salvar
                          </button>
                          <button
                            onClick={cancelEdit}
                            className="btn btn-secondary btn-sm"
                          >
                            ❌ Cancelar
                          </button>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="message-content">
                          <p>{message.content}</p>
                          <span className="message-date">
                            {formatDate(message.created_at)}
                          </span>
                        </div>
                        <div className="message-actions">
                          <button
                            onClick={() => startEdit(message)}
                            className="btn btn-edit btn-sm"
                            title="Editar"
                          >
                            ✏️
                          </button>
                          <button
                            onClick={() => deleteMessage(message.id)}
                            className="btn btn-delete btn-sm"
                            title="Deletar"
                          >
                            🗑️
                          </button>
                        </div>
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </main>
    </div>
  );
}
