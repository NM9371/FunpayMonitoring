-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     telegram_id BIGINT UNIQUE NOT NULL,
                                     username TEXT,
                                     created_at TIMESTAMP DEFAULT NOW()
    );

-- Таблица подписок на лоты
CREATE TABLE IF NOT EXISTS subscriptions (
                                             id SERIAL PRIMARY KEY,
                                             user_id BIGINT NOT NULL,
                                             lot_name TEXT NOT NULL,
                                             min_price NUMERIC(12,2) NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );

-- Таблица истории цен лотов
CREATE TABLE IF NOT EXISTS price_history (
                                             id SERIAL PRIMARY KEY,
                                             lot_name TEXT NOT NULL,
                                             price NUMERIC(12,2) NOT NULL,
    checked_at TIMESTAMP DEFAULT NOW()
    );

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_subscriptions_lot_name ON subscriptions(lot_name);
CREATE INDEX IF NOT EXISTS idx_price_history_lot_name ON price_history(lot_name);