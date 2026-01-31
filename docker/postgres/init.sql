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
    category TEXT NOT NULL,
    lot_name TEXT NOT NULL,
    min_price NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица истории цен лотов
CREATE TABLE IF NOT EXISTS price_history (
    id SERIAL PRIMARY KEY,
    category TEXT NOT NULL,
    lot_name TEXT NOT NULL,
    price NUMERIC(12,2) NOT NULL,
    checked_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_price_history_lot_name ON price_history(lot_name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_unique
    ON subscriptions(user_id, category, lot_name);
