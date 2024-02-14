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

CREATE TABLE carteiras (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    valor INTEGER NOT NULL,
    ultimas_transacoes json[] NULL
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
-- test
    INSERT INTO carteiras(cliente_id, valor)
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
    UPDATE carteiras
    SET valor = valor + p2,
        ultimas_transacoes = (
        SELECT json_build_object(
         'valor', p2,
         'tipo', 'c',
         'descricao', p3,
         'realizada_em', now()
        ) || ultimas_transacoes)[:10]
    WHERE cliente_id = p1
    RETURNING valor INTO sv;
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
    UPDATE carteiras
    SET valor = valor - p3,
        ultimas_transacoes = (
            SELECT json_build_object(
               'valor', p3,
               'tipo', 'd',
               'descricao', p4,
               'realizada_em', now()
           ) || ultimas_transacoes)[:10]
    WHERE cliente_id = p1 AND valor - p3 > p2
        RETURNING valor INTO sv;

    RETURN sv;
END;
$$ LANGUAGE plpgsql;
