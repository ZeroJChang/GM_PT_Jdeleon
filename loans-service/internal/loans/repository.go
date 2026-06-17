package loans

import (
"context"
"database/sql"
"errors"
)

var (
ErrLoanNotFound     = errors.New("loan not found")
ErrActiveLoanExists = errors.New("user already has an active loan for this book")
)

type Repository struct {
db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
return &Repository{db: db}
}

func (r *Repository) HasActiveLoan(ctx context.Context, userID string, bookID string) (bool, error) {
var count int

err := r.db.QueryRowContext(
ctx,
`
SELECT COUNT(1)
FROM loans
WHERE user_id = $1
  AND book_id = $2
  AND status = $3
`,
userID,
bookID,
StatusActive,
).Scan(&count)

if err != nil {
return false, err
}

return count > 0, nil
}

func (r *Repository) Create(ctx context.Context, loanID string, userID string, bookID string) (*Loan, error) {
row := r.db.QueryRowContext(
ctx,
`
INSERT INTO loans (id, user_id, book_id, status)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, book_id, status, loan_date, return_date, created_at, updated_at
`,
loanID,
userID,
bookID,
StatusActive,
)

return scanLoan(row)
}

func (r *Repository) ReturnLoan(ctx context.Context, loanID string) (*Loan, error) {
row := r.db.QueryRowContext(
ctx,
`
UPDATE loans
SET status = $1,
    return_date = NOW(),
    updated_at = NOW()
WHERE id = $2
  AND status = $3
RETURNING id, user_id, book_id, status, loan_date, return_date, created_at, updated_at
`,
StatusReturned,
loanID,
StatusActive,
)

loan, err := scanLoan(row)
if errors.Is(err, sql.ErrNoRows) {
return nil, ErrLoanNotFound
}

return loan, err
}

func (r *Repository) FindActiveByUser(ctx context.Context, userID string) ([]Loan, error) {
rows, err := r.db.QueryContext(
ctx,
`
SELECT id, user_id, book_id, status, loan_date, return_date, created_at, updated_at
FROM loans
WHERE user_id = $1
  AND status = $2
ORDER BY loan_date DESC
`,
userID,
StatusActive,
)
if err != nil {
return nil, err
}
defer rows.Close()

return scanLoans(rows)
}

func (r *Repository) FindHistory(ctx context.Context) ([]Loan, error) {
rows, err := r.db.QueryContext(
ctx,
`
SELECT id, user_id, book_id, status, loan_date, return_date, created_at, updated_at
FROM loans
ORDER BY loan_date DESC
`,
)
if err != nil {
return nil, err
}
defer rows.Close()

return scanLoans(rows)
}

type scanner interface {
Scan(dest ...any) error
}

func scanLoan(row scanner) (*Loan, error) {
var loan Loan

err := row.Scan(
&loan.ID,
&loan.UserID,
&loan.BookID,
&loan.Status,
&loan.LoanDate,
&loan.ReturnDate,
&loan.CreatedAt,
&loan.UpdatedAt,
)

if err != nil {
return nil, err
}

return &loan, nil
}

func scanLoans(rows *sql.Rows) ([]Loan, error) {
loans := make([]Loan, 0)

for rows.Next() {
loan, err := scanLoan(rows)
if err != nil {
return nil, err
}

loans = append(loans, *loan)
}

if err := rows.Err(); err != nil {
return nil, err
}

return loans, nil
}
