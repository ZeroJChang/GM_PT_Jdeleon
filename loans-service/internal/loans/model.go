package loans

import "time"

const (
StatusActive   = "ACTIVE"
StatusReturned = "RETURNED"
)

type Loan struct {
ID         string     `json:"id"`
UserID     string     `json:"userId"`
BookID     string     `json:"bookId"`
Status     string     `json:"status"`
LoanDate   time.Time  `json:"loanDate"`
ReturnDate *time.Time `json:"returnDate,omitempty"`
CreatedAt  time.Time  `json:"createdAt"`
UpdatedAt  time.Time  `json:"updatedAt"`
}

type RegisterLoanRequest struct {
UserID string `json:"userId"`
BookID string `json:"bookId"`
}
