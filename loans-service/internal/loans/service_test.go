package loans

import (
"context"
"errors"
"testing"
)

type fakeRepository struct {
hasActiveLoan bool
hasActiveErr  error
createErr     error
returnErr     error

createdLoan  *Loan
returnedLoan *Loan
}

func (r *fakeRepository) HasActiveLoan(ctx context.Context, userID string, bookID string) (bool, error) {
return r.hasActiveLoan, r.hasActiveErr
}

func (r *fakeRepository) Create(ctx context.Context, loanID string, userID string, bookID string) (*Loan, error) {
if r.createErr != nil {
return nil, r.createErr
}

r.createdLoan = &Loan{
ID:     loanID,
UserID: userID,
BookID: bookID,
Status: StatusActive,
}

return r.createdLoan, nil
}

func (r *fakeRepository) ReturnLoan(ctx context.Context, loanID string) (*Loan, error) {
if r.returnErr != nil {
return nil, r.returnErr
}

r.returnedLoan = &Loan{
ID:     loanID,
UserID: "user-1",
BookID: "book-1",
Status: StatusReturned,
}

return r.returnedLoan, nil
}

func (r *fakeRepository) FindActiveByUser(ctx context.Context, userID string) ([]Loan, error) {
return []Loan{
{
ID:     "loan-1",
UserID: userID,
BookID: "book-1",
Status: StatusActive,
},
}, nil
}

func (r *fakeRepository) FindHistory(ctx context.Context) ([]Loan, error) {
return []Loan{
{
ID:     "loan-1",
UserID: "user-1",
BookID: "book-1",
Status: StatusReturned,
},
}, nil
}

type fakeLibraryClient struct {
reserveCalled bool
releaseCalled bool

reserveErr error
releaseErr error
}

func (c *fakeLibraryClient) ReserveBook(ctx context.Context, bookID string) error {
c.reserveCalled = true
return c.reserveErr
}

func (c *fakeLibraryClient) ReleaseBook(ctx context.Context, bookID string) error {
c.releaseCalled = true
return c.releaseErr
}

func TestRegisterLoanCreatesLoanAndReservesBook(t *testing.T) {
repository := &fakeRepository{}
libraryClient := &fakeLibraryClient{}

service := NewService(repository, libraryClient)

loan, err := service.RegisterLoan(context.Background(), RegisterLoanRequest{
UserID: "user-1",
BookID: "book-1",
})

if err != nil {
t.Fatalf("expected no error, got %v", err)
}

if loan == nil {
t.Fatal("expected loan, got nil")
}

if loan.UserID != "user-1" {
t.Fatalf("expected user-1, got %s", loan.UserID)
}

if loan.BookID != "book-1" {
t.Fatalf("expected book-1, got %s", loan.BookID)
}

if loan.Status != StatusActive {
t.Fatalf("expected ACTIVE status, got %s", loan.Status)
}

if !libraryClient.reserveCalled {
t.Fatal("expected ReserveBook to be called")
}
}

func TestRegisterLoanReturnsErrorWhenActiveLoanExists(t *testing.T) {
repository := &fakeRepository{
hasActiveLoan: true,
}
libraryClient := &fakeLibraryClient{}

service := NewService(repository, libraryClient)

loan, err := service.RegisterLoan(context.Background(), RegisterLoanRequest{
UserID: "user-1",
BookID: "book-1",
})

if !errors.Is(err, ErrActiveLoanExists) {
t.Fatalf("expected ErrActiveLoanExists, got %v", err)
}

if loan != nil {
t.Fatal("expected nil loan")
}

if libraryClient.reserveCalled {
t.Fatal("expected ReserveBook not to be called")
}
}

func TestRegisterLoanReleasesBookWhenCreateFails(t *testing.T) {
repository := &fakeRepository{
createErr: errors.New("database error"),
}
libraryClient := &fakeLibraryClient{}

service := NewService(repository, libraryClient)

loan, err := service.RegisterLoan(context.Background(), RegisterLoanRequest{
UserID: "user-1",
BookID: "book-1",
})

if err == nil {
t.Fatal("expected error")
}

if loan != nil {
t.Fatal("expected nil loan")
}

if !libraryClient.reserveCalled {
t.Fatal("expected ReserveBook to be called")
}

if !libraryClient.releaseCalled {
t.Fatal("expected ReleaseBook to be called as compensation")
}
}

func TestReturnLoanReturnsLoanAndReleasesBook(t *testing.T) {
repository := &fakeRepository{}
libraryClient := &fakeLibraryClient{}

service := NewService(repository, libraryClient)

loan, err := service.ReturnLoan(context.Background(), "loan-1")

if err != nil {
t.Fatalf("expected no error, got %v", err)
}

if loan == nil {
t.Fatal("expected loan, got nil")
}

if loan.Status != StatusReturned {
t.Fatalf("expected RETURNED status, got %s", loan.Status)
}

if !libraryClient.releaseCalled {
t.Fatal("expected ReleaseBook to be called")
}
}
