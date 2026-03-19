CREATE DATABASE IF NOT EXISTS shellbin;
USE shellbin;

CREATE TABLE IF NOT EXISTS pastes (
  hash VARCHAR(8) NOT NULL PRIMARY KEY,
  content TEXT NOT NULL,
  submission_date DATE NOT NULL,
  last_used DATE NOT NULL
);
