CREATE TABLE IF NOT EXISTS comment(
  id SERIAL PRIMARY KEY,
  text TEXT NOT NULL,
  text_search tsvector GENERATED ALWAYS AS (to_tsvector('simple', text)) STORED
);

CREATE INDEX IF NOT EXISTS idx_text_search on comment USING gin(text_search);

CREATE TABLE IF NOT EXISTS comment_closure_tree(
  parent_id INTEGER NOT NULL,
  child_id INTEGER NOT NULL,
  depth INTEGER NOT NULL,
  PRIMARY KEY (parent_id, child_id),
  FOREIGN KEY (parent_id) REFERENCES comment(id) ON DELETE CASCADE,
  FOREIGN KEY (child_id) REFERENCES comment(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_parent ON comment_closure_tree(parent_id);
CREATE INDEX IF NOT EXISTS idx_child ON comment_closure_tree(child_id);
