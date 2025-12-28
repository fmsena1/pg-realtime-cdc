-- ===============================
-- Tabela
-- ===============================
CREATE TABLE IF NOT EXISTS messages (
  id SERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT now()
);

-- ===============================
-- Publication (idempotente)
-- ===============================
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_publication
    WHERE pubname = 'go_realtime_pub'
  ) THEN
    CREATE PUBLICATION go_realtime_pub
    FOR TABLE messages;
  END IF;
END
$$;

-- ===============================
-- Permissão de replicação
-- ===============================
ALTER ROLE realtime WITH REPLICATION;

-- ===============================
-- Logical Replication Slot (opcional mas recomendado)
-- ⚠️ Só cria se não existir
-- ===============================
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_replication_slots
    WHERE slot_name = 'go_realtime_slot'
  ) THEN
    PERFORM pg_create_logical_replication_slot(
      'go_realtime_slot',
      'pgoutput'
    );
  END IF;
END
$$;
