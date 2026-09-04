CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    username VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- INSERT INTO users (username, email, password, avatar) VALUES ('pippo', 'pippo@p.p', 'abcd', 'avatar_deity_man_02.png');
-- INSERT INTO users (username, email, password, avatar) VALUES ('pluto', 'pluto@p.p', 'abcd', 'avatar_warrior_halfork_woman_01.png');

CREATE TABLE IF NOT EXISTS chats(
    id SERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    location VARCHAR(30) NOT NULL,
    description VARCHAR(255) NOT NULL,
    is_open BOOLEAN DEFAULT true NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_chat_user FOREIGN KEY (user_id) REFERENCES users_test(id)
);

-- INSERT INTO chats (name, slug, location, description, is_open, user_id) VALUES ('example', 'example', 'la piazza', 'Suspendisse molestie ultricies neque sit amet mollis. Donec euismod libero quis felis eleifend blandit pharetra in libero...', true, 1);

CREATE TABLE IF NOT EXISTS messages(
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    text TEXT NOT NULL,
    chat_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_message_user FOREIGN KEY (user_id) REFERENCES users_test(id) ON DELETE CASCADE,
    CONSTRAINT fk_message_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);
