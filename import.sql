--CREATE ROLE root WITH SUPERUSER LOGIN PASSWORD 'pass'; -- para arrumar um log insuportavel
--CREATE DATABASE root; -- para arrumar um log insuportavel

CREATE TABLE clientes (
  id SERIAL PRIMARY KEY,
  nome VARCHAR (50) NOT NULL,
  limite INTEGER NOT NULL
);

CREATE TABLE transacoes (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    valor INTEGER NOT NULL,
    tipo CHAR(1) NOT NULL,
    descricao VARCHAR(10) NOT NULL,
    realizada_em TIMESTAMP NOT NULL,
    CONSTRAINT fk_clientes_transacoes_id
        FOREIGN KEY (cliente_id) REFERENCES clientes(id),
    UNIQUE (cliente_id, valor, tipo, descricao, realizada_em)
);

CREATE TABLE carteiras (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    valor INTEGER NOT NULL,
    ultimas_transacoes json[] NULL
);

DO $$
BEGIN
    INSERT INTO clientes (nome, limite)
    VALUES
        ('cliente 1', 1000 * 100),
        ('cliente 2', 800 * 100),
        ('cliente 3', 10000 * 100),
        ('cliente 4', 100000 * 100),
        ('cliente 5', 5000 * 100);
    INSERT INTO carteiras(cliente_id, valor)
        SELECT id, 0 FROM clientes;
END;
$$;


CREATE OR REPLACE FUNCTION notify_new_transaction()
RETURNS TRIGGER AS
$$
BEGIN
    PERFORM pg_notify(
        'transaction_added',
        json_build_object(
            'transaction', NEW.ultimas_transacoes[1],
            'customer_id', NEW.cliente_id
        )::TEXT
    );
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER carteiras_update_trigger
    AFTER UPDATE ON carteiras
    FOR EACH ROW
    EXECUTE FUNCTION notify_new_transaction();