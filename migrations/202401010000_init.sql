-- Создание таблицы для хранения запросов на сканирование
CREATE TABLE scan_requests (
                               id SERIAL PRIMARY KEY,
                               ip_address VARCHAR(45) NOT NULL,
                               ports VARCHAR(255) NOT NULL,
                               created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для хранения результатов сканирования
CREATE TABLE scan_results (
                              id SERIAL PRIMARY KEY,
                              request_id INTEGER NOT NULL REFERENCES scan_requests(id) ON DELETE CASCADE,
                              port INTEGER NOT NULL,
                              is_open BOOLEAN NOT NULL,
                              scanned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска результатов по ID запроса
CREATE INDEX idx_scan_results_request_id ON scan_results(request_id);

-- Индекс для быстрого поиска открытых портов
CREATE INDEX idx_scan_results_is_open ON scan_results(is_open) WHERE is_open = true;