CREATE TABLE IF NOT EXISTS loans (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    book_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    loan_date TIMESTAMP NOT NULL DEFAULT NOW(),
    return_date TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_loans_user_id ON loans(user_id);
CREATE INDEX IF NOT EXISTS idx_loans_book_id ON loans(book_id);
CREATE INDEX IF NOT EXISTS idx_loans_status ON loans(status);
