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
    realizada_em TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_clientes_transacoes_id
        FOREIGN KEY (cliente_id) REFERENCES clientes(id)
);

CREATE INDEX idx_transacoes_id_desc ON transacoes(id desc);

CREATE TABLE saldos (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    valor INTEGER NOT NULL,
    CONSTRAINT fk_clientes_saldos_id
        FOREIGN KEY (cliente_id) REFERENCES clientes(id)
);

CREATE UNIQUE INDEX idx_saldos_cliente_id ON saldos (cliente_id) include (valor);

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


CREATE OR REPLACE FUNCTION c(
    p1 INT, --customer.id
    p2 INT, --value
    p3 VARCHAR(10) -- description
)
RETURNS INT AS $$
DECLARE
    sv INT;
BEGIN
    UPDATE saldos SET valor = valor + p2
    WHERE cliente_id = p1
        RETURNING valor INTO sv;

    INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em)
    VALUES (p1, p2, 'c', p3, now());

    RETURN sv;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION d(
    p1 INT, --customer.id
    p2 INT, --limit
    p3 INT, --value
    p4 VARCHAR(10) --description
)
RETURNS INT AS $$
DECLARE
    sv INT;
BEGIN
    UPDATE saldos SET valor = valor - p3
    WHERE cliente_id = p1 AND valor - p3 > p2
    RETURNING valor INTO sv;

    IF sv IS NULL THEN
           RETURN NULL;
    END IF;

    INSERT INTO transacoes (cliente_id, valor, tipo, descricao, realizada_em)
    VALUES (p1, p3, 'd', p4, now());

    RETURN sv;
END;
$$ LANGUAGE plpgsql;
