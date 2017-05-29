DROP DATABASE IF EXISTS docsensesearch;
CREATE DATABASE docsensesearch
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE docsensesearch;

CREATE TABLE files (
  id       SERIAL,
  filename VARCHAR(1023) NOT NULL,
  sc_id    TEXT          NOT NULL,
  link     VARCHAR(1023) NOT NULL,
  sp_id    INTEGER       NOT NULL,
  sp_list  INTEGER       NOT NULL,
  position INTEGER       NOT NULL,
  year     INTEGER       NOT NULL,
  date     DATETIME      NOT NULL
);

CREATE TABLE sp_lists (
  id             SERIAL,
  link           TEXT,
  sp_id          TEXT,
  last_migration DATETIME
);
