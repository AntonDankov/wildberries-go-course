CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    amount INT NOT NULL CHECK (amount >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS item_history (
    id SERIAL PRIMARY KEY,
    item_id BIGINT NOT NULL,
    name VARCHAR(255),
    price DECIMAL(10, 2),
    amount INT,
    action INT NOT NULL,
    changed_by BIGINT NOT NULL REFERENCES users(id),
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION log_item_change()
RETURNS TRIGGER AS $$
DECLARE
    user_id BIGINT;
    changed_name VARCHAR(255);
    changed_price DECIMAL(10, 2);
    changed_amount INT;
BEGIN
    user_id := current_setting('app.current_user_id', true)::BIGINT;
    
    IF user_id IS NULL THEN
        RAISE EXCEPTION 'app.current_user_id must be set';
    END IF;

    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history (
            item_id, name, price, amount, action, changed_by
        ) VALUES (
            NEW.id, NEW.name, NEW.price, NEW.amount, 0, user_id
        );
        RETURN NEW;
        
    ELSIF TG_OP = 'UPDATE' THEN
        changed_name := NULL;
        changed_price := NULL;
        changed_amount := NULL;
        
        IF OLD.name IS DISTINCT FROM NEW.name THEN
            changed_name := NEW.name;
        END IF;
        
        IF OLD.price IS DISTINCT FROM NEW.price THEN
            changed_price := NEW.price;
        END IF;
        
        IF OLD.amount IS DISTINCT FROM NEW.amount THEN
            changed_amount := NEW.amount;
        END IF;
        
        IF changed_name IS NOT NULL OR changed_price IS NOT NULL OR changed_amount IS NOT NULL THEN
            INSERT INTO item_history (
                item_id, name, price, amount, action, changed_by
            ) VALUES (
                NEW.id, changed_name, changed_price, changed_amount, 1, user_id
            );
        END IF;
        
        RETURN NEW;
        
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history (
            item_id, name, price, amount, action, changed_by
        ) VALUES (
            OLD.id, OLD.name, OLD.price, OLD.amount, 2, user_id
        );
        RETURN OLD;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'item_change_trigger'
    ) THEN
        CREATE TRIGGER item_change_trigger
        AFTER INSERT OR UPDATE OR DELETE ON items
        FOR EACH ROW EXECUTE FUNCTION log_item_change();
    END IF;
END;
$$;
