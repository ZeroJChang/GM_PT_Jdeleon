package loans

import (
"context"

"github.com/google/uuid"
)

type LibraryBookClient interface {
ReserveBook(ctx context.Context, bookID string) error
ReleaseBook(ctx context.Context, bookID string) error
}

type Service struct {
repository    *Repository
libraryClient LibraryBookClient
}

func NewService(repository *Repository, libraryClient LibraryBookClient) *Service {
return &Service{
repository:    repository,
libraryClient: libraryClient,
}
}

func (s *Service) RegisterLoan(ctx context.Context, request RegisterLoanRequest) (*Loan, error) {
hasActiveLoan, err := s.repository.HasActiveLoan(ctx, request.UserID, request.BookID)
if err != nil {
return nil, err
}

if hasActiveLoan {
return nil, ErrActiveLoanExists
}

if err := s.libraryClient.ReserveBook(ctx, request.BookID); err != nil {
return nil, err
}

loan, err := s.repository.Create(ctx, uuid.NewString(), request.UserID, request.BookID)
if err != nil {
_ = s.libraryClient.ReleaseBook(ctx, request.BookID)
return nil, err
}

return loan, nil
}

func (s *Service) ReturnLoan(ctx context.Context, loanID string) (*Loan, error) {
loan, err := s.repository.ReturnLoan(ctx, loanID)
if err != nil {
return nil, err
}

if err := s.libraryClient.ReleaseBook(ctx, loan.BookID); err != nil {
return nil, err
}

return loan, nil
}

func (s *Service) FindActiveByUser(ctx context.Context, userID string) ([]Loan, error) {
return s.repository.FindActiveByUser(ctx, userID)
}

func (s *Service) FindHistory(ctx context.Context) ([]Loan, error) {
return s.repository.FindHistory(ctx)
}
