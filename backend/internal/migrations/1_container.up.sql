CREATE TABLE containers (
    ip_address TEXT PRIMARY KEY,
    is_reachable BOOLEAN,
    last_ping TIMESTAMP WITHOUT TIME ZONE
);