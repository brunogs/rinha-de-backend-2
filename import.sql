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
    realizada_em TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_clientes_transacoes_id
        FOREIGN KEY (cliente_id) REFERENCES clientes(id)
);

CREATE TABLE saldos (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    valor INTEGER NOT NULL,
    CONSTRAINT fk_clientes_saldos_id
        FOREIGN KEY (cliente_id) REFERENCES clientes(id)
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
    INSERT INTO saldos (cliente_id, valor)
        SELECT id, 0 FROM clientes;
END;
$$;