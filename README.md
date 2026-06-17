\# Library Technical Test



Backend technical test for a library system implemented with two independent services:



\* \*\*Library Service\*\*: Main API exposed to the client.

\* \*\*Loans Service\*\*: Dedicated loan management service implemented in Go.



The system allows users to authenticate, manage books and users, register loans, return loans, list active loans, and view loan history. The services communicate through HTTP and each service owns its own PostgreSQL database.



\---



\## Table of Contents



\* \[Architecture](#architecture)

\* \[Technology Stack](#technology-stack)

\* \[Project Structure](#project-structure)

\* \[Main Decisions](#main-decisions)

\* \[Requirements Covered](#requirements-covered)

\* \[Environment Variables](#environment-variables)

\* \[How to Run with Docker](#how-to-run-with-docker)

\* \[Demo Data](#demo-data)

\* \[API Documentation](#api-documentation)

\* \[Main Flow](#main-flow)

\* \[Tests](#tests)

\* \[Useful Commands](#useful-commands)

\* \[Troubleshooting](#troubleshooting)

\* \[Trade-offs](#trade-offs)



\---



\## Architecture



The solution is composed of two backend services and two independent PostgreSQL databases.



```mermaid

flowchart LR

&#x20;   Client\[Client / API Consumer]



&#x20;   subgraph ServiceA\[Service A - library-service]

&#x20;       Nest\[NestJS API]

&#x20;       Auth\[Auth JWT + Roles]

&#x20;       Books\[Books Module]

&#x20;       Users\[Users Module]

&#x20;       Internal\[Internal Books API]

&#x20;       LoansProxy\[Loans Module / HTTP Client]

&#x20;   end



&#x20;   subgraph DB1\[(library-db)]

&#x20;       UsersTable\[users]

&#x20;       BooksTable\[books]

&#x20;   end



&#x20;   subgraph ServiceB\[Service B - loans-service]

&#x20;       GoAPI\[Go Chi API]

&#x20;       LoanBusiness\[Loans Business Logic]

&#x20;       LibraryClient\[HTTP Client to Library Service]

&#x20;   end



&#x20;   subgraph DB2\[(loans-db)]

&#x20;       LoansTable\[loans]

&#x20;   end



&#x20;   Client --> Nest

&#x20;   Nest --> Auth

&#x20;   Nest --> Books

&#x20;   Nest --> Users

&#x20;   Nest --> LoansProxy



&#x20;   Books --> DB1

&#x20;   Users --> DB1



&#x20;   LoansProxy -->|HTTP| GoAPI

&#x20;   GoAPI --> LoanBusiness

&#x20;   LoanBusiness --> DB2

&#x20;   LoanBusiness --> LibraryClient

&#x20;   LibraryClient -->|HTTP Internal API| Internal

&#x20;   Internal --> DB1

```



\---



\## Service Communication



The client communicates only with \*\*library-service\*\*.



When a loan is created, the flow is:



```mermaid

sequenceDiagram

&#x20;   participant Client

&#x20;   participant Library as library-service

&#x20;   participant Loans as loans-service

&#x20;   participant DBLoans as loans-db

&#x20;   participant DBLibrary as library-db



&#x20;   Client->>Library: POST /loans with JWT

&#x20;   Library->>Loans: POST /loans

&#x20;   Loans->>Library: POST /internal/books/:id/reserve

&#x20;   Library->>DBLibrary: Decrease availableCopies

&#x20;   Library-->>Loans: Book reserved

&#x20;   Loans->>DBLoans: Create ACTIVE loan

&#x20;   Loans-->>Library: Loan created

&#x20;   Library-->>Client: 201 Created

```



When a loan is returned:



```mermaid

sequenceDiagram

&#x20;   participant Client

&#x20;   participant Library as library-service

&#x20;   participant Loans as loans-service

&#x20;   participant DBLoans as loans-db

&#x20;   participant DBLibrary as library-db



&#x20;   Client->>Library: POST /loans/:id/return with JWT

&#x20;   Library->>Loans: POST /loans/:id/return

&#x20;   Loans->>DBLoans: Mark loan as RETURNED

&#x20;   Loans->>Library: POST /internal/books/:id/release

&#x20;   Library->>DBLibrary: Increase availableCopies

&#x20;   Loans-->>Library: Loan returned

&#x20;   Library-->>Client: 200 OK

```



\---



\## Technology Stack



\### Service A: `library-service`



\* Node.js

\* NestJS

\* Prisma ORM

\* PostgreSQL

\* JWT authentication

\* Role-based authorization

\* Class Validator

\* Docker



\### Service B: `loans-service`



\* Go

\* Chi router

\* PostgreSQL

\* `database/sql`

\* `pgx` driver

\* Docker



\### Infrastructure



\* Docker Compose

\* Two PostgreSQL databases

\* Independent containers per service



\---



\## Project Structure



```text

library-technical-test/

&#x20; docker-compose.yml

&#x20; README.md

&#x20; .gitignore



&#x20; library-service/

&#x20;   Dockerfile

&#x20;   .dockerignore

&#x20;   .env.example

&#x20;   prisma/

&#x20;     schema.prisma

&#x20;     seed.cjs

&#x20;     migrations/

&#x20;   src/

&#x20;     auth/

&#x20;     books/

&#x20;     database/

&#x20;     internal/

&#x20;     loans/

&#x20;     users/



&#x20; loans-service/

&#x20;   Dockerfile

&#x20;   .dockerignore

&#x20;   .env.example

&#x20;   go.mod

&#x20;   go.sum

&#x20;   cmd/

&#x20;     server/

&#x20;       main.go

&#x20;   migrations/

&#x20;     001\_init.sql

&#x20;   internal/

&#x20;     loans/

&#x20;     platform/

&#x20;       database/

&#x20;       httpclient/

```



\---



\## Main Decisions



\### 1. Library Service as the main public API



The client interacts with `library-service`. This service exposes authentication, users, books, and loan endpoints.



The loan endpoints in `library-service` act as a gateway to `loans-service`.



\### 2. Independent databases



Each service owns its data:



\* `library-db` stores users and books.

\* `loans-db` stores loans.



This avoids direct database coupling between services.



\### 3. HTTP communication between services



The services communicate through REST over HTTP.



\* `library-service` calls `loans-service` to register and return loans.

\* `loans-service` calls internal endpoints from `library-service` to reserve or release book copies.



\### 4. Internal API key



Internal endpoints are protected using:



```http

x-internal-api-key: local\_internal\_key

```



This prevents regular clients from directly reserving or releasing book copies without authorization.



\### 5. Compensation instead of distributed transactions



There is no distributed transaction between the two databases.



If `loans-service` reserves a book but fails to create the loan, it calls `release` in `library-service` as a compensation mechanism.



\---



\## Requirements Covered



\* Books CRUD.

\* Users CRUD.

\* JWT authentication.

\* Roles: `ADMIN` and `USER`.

\* Book filters and pagination.

\* Register loans.

\* Return loans.

\* List active loans by user.

\* List loan history.

\* HTTP communication between services.

\* Independent persistence per service.

\* Docker Compose for both services and databases.

\* Seed data for easy evaluation.

\* Unit tests for both services.

\* README documentation with architecture, flow, endpoints, decisions and execution instructions.



\---



\## Environment Variables



\### `library-service/.env.example`



```env

PORT=3000

DATABASE\_URL=postgresql://postgres:postgres@localhost:5433/library\_db?schema=public

JWT\_SECRET=local\_super\_secret\_key

JWT\_EXPIRES\_IN=1d

LOANS\_SERVICE\_URL=http://localhost:8080

INTERNAL\_API\_KEY=local\_internal\_key

```



\### `loans-service/.env.example`



```env

PORT=8080

DATABASE\_URL=postgres://postgres:postgres@localhost:5434/loans\_db?sslmode=disable

LIBRARY\_SERVICE\_URL=http://localhost:3000

INTERNAL\_API\_KEY=local\_internal\_key

```



When running with Docker Compose, the environment variables are already defined in `docker-compose.yml`.



\---



\## How to Run with Docker



\### Prerequisites



\* Docker Desktop installed.

\* Docker Compose available.

\* Ports `3000`, `8080`, `5433`, and `5434` available.



\### Run the full system



From the root folder:



```powershell

docker compose up --build

```



This command starts:



\* `library-service`

\* `loans-service`

\* `library-db`

\* `loans-db`



The `library-service` container runs automatically:



```text

npx prisma migrate deploy

npm run seed

node dist/src/main.js

```



This means the database schema is applied and demo data is created automatically.



\### Validate services



Open another terminal and run:



```powershell

Invoke-RestMethod -Uri "http://localhost:3000" -Method GET

```



Expected response:



```text

Hello World!

```



Validate Go service health:



```powershell

Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET

```



Expected response:



```json

{

&#x20; "status": "ok",

&#x20; "service": "loans-service",

&#x20; "database": "ok"

}

```



\---



\## Demo Data



The seed automatically creates:



\### Admin User



```text

Email: admin@test.com

Password: 123456

Role: ADMIN

```



\### Normal User



```text

Email: user@test.com

Password: 123456

Role: USER

```



\### Demo Book



```text

Title: Clean Code

Author: Robert C. Martin

ISBN: 9780132350884

Year: 2008

Genre: Software Engineering

Total copies: 5

Available copies: 5

```



\---



\# API Documentation



Base URL for the main public API:



```text

http://localhost:3000

```



Base URL for the loans service:



```text

http://localhost:8080

```



The client should normally consume only `library-service`.



\---



\## Authentication



\### Login



```http

POST /auth/login

```



Request:



```json

{

&#x20; "email": "admin@test.com",

&#x20; "password": "123456"

}

```



Response:



```json

{

&#x20; "accessToken": "jwt\_token\_here",

&#x20; "user": {

&#x20;   "id": "e11067e4-7927-442b-89ea-533cf5564609",

&#x20;   "name": "Admin User",

&#x20;   "email": "admin@test.com",

&#x20;   "role": "ADMIN"

&#x20; }

}

```



PowerShell example:



```powershell

$adminResponse = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/auth/login" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Body '{

&#x20;   "email": "admin@test.com",

&#x20;   "password": "123456"

&#x20; }'



$adminToken = $adminResponse.accessToken

```



\---



\## Books Endpoints



\### Create Book



Requires role: `ADMIN`



```http

POST /books

```



Request:



```json

{

&#x20; "title": "Clean Architecture",

&#x20; "author": "Robert C. Martin",

&#x20; "isbn": "9780134494166",

&#x20; "year": 2017,

&#x20; "genre": "Software Engineering",

&#x20; "totalCopies": 3

}

```



Response:



```json

{

&#x20; "id": "book-uuid",

&#x20; "title": "Clean Architecture",

&#x20; "author": "Robert C. Martin",

&#x20; "isbn": "9780134494166",

&#x20; "year": 2017,

&#x20; "genre": "Software Engineering",

&#x20; "totalCopies": 3,

&#x20; "availableCopies": 3,

&#x20; "createdAt": "2026-06-17T00:00:00.000Z",

&#x20; "updatedAt": "2026-06-17T00:00:00.000Z"

}

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/books" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" } `

&#x20; -Body '{

&#x20;   "title": "Clean Architecture",

&#x20;   "author": "Robert C. Martin",

&#x20;   "isbn": "9780134494166",

&#x20;   "year": 2017,

&#x20;   "genre": "Software Engineering",

&#x20;   "totalCopies": 3

&#x20; }'

```



\---



\### List Books



Requires authentication.



```http

GET /books

```



Query parameters:



| Name        |    Type | Required | Description                           |

| ----------- | ------: | -------: | ------------------------------------- |

| `author`    |  string |       No | Filter by author                      |

| `genre`     |  string |       No | Filter by genre                       |

| `available` | boolean |       No | Filter available or unavailable books |

| `page`      |  number |       No | Page number                           |

| `limit`     |  number |       No | Items per page                        |



Example:



```http

GET /books?author=Robert\&available=true\&page=1\&limit=10

```



Response:



```json

{

&#x20; "items": \[

&#x20;   {

&#x20;     "id": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",

&#x20;     "title": "Clean Code",

&#x20;     "author": "Robert C. Martin",

&#x20;     "isbn": "9780132350884",

&#x20;     "year": 2008,

&#x20;     "genre": "Software Engineering",

&#x20;     "totalCopies": 5,

&#x20;     "availableCopies": 5

&#x20;   }

&#x20; ],

&#x20; "meta": {

&#x20;   "total": 1,

&#x20;   "page": 1,

&#x20;   "limit": 10,

&#x20;   "totalPages": 1

&#x20; }

}

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/books" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" }

```



\---



\### Get Book by ID



Requires authentication.



```http

GET /books/:id

```



Example:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/books/117f2aaf-61ab-4b49-b34a-7b331f6947a8" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" }

```



\---



\### Update Book



Requires role: `ADMIN`



```http

PATCH /books/:id

```



Request:



```json

{

&#x20; "title": "Updated title",

&#x20; "totalCopies": 10

}

```



\---



\### Delete Book



Requires role: `ADMIN`



```http

DELETE /books/:id

```



Response:



```json

{

&#x20; "message": "Book deleted successfully"

}

```



\---



\## Users Endpoints



All users endpoints require role: `ADMIN`.



\### Create User



```http

POST /users

```



Request:



```json

{

&#x20; "name": "Normal User",

&#x20; "email": "user@test.com",

&#x20; "password": "123456",

&#x20; "role": "USER"

}

```



Response:



```json

{

&#x20; "id": "user-uuid",

&#x20; "name": "Normal User",

&#x20; "email": "user@test.com",

&#x20; "role": "USER",

&#x20; "createdAt": "2026-06-17T00:00:00.000Z",

&#x20; "updatedAt": "2026-06-17T00:00:00.000Z"

}

```



The password hash is never returned.



\---



\### List Users



```http

GET /users

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/users" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" }

```



\---



\### Get User by ID



```http

GET /users/:id

```



\---



\### Update User



```http

PATCH /users/:id

```



Request:



```json

{

&#x20; "name": "Updated User",

&#x20; "role": "ADMIN"

}

```



\---



\### Delete User



```http

DELETE /users/:id

```



Response:



```json

{

&#x20; "message": "User deleted successfully"

}

```



\---



\## Loans Endpoints Through Library Service



These are the endpoints that the client should use.



\### Register Loan



Requires authentication.



```http

POST /loans

```



Request as normal user:



```json

{

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"

}

```



Request as admin for another user:



```json

{

&#x20; "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"

}

```



Response:



```json

{

&#x20; "id": "loan-uuid",

&#x20; "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",

&#x20; "status": "ACTIVE",

&#x20; "loanDate": "2026-06-17T02:08:12.358975Z",

&#x20; "createdAt": "2026-06-17T02:08:12.358975Z",

&#x20; "updatedAt": "2026-06-17T02:08:12.358975Z"

}

```



PowerShell:



```powershell

$userResponse = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/auth/login" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Body '{

&#x20;   "email": "user@test.com",

&#x20;   "password": "123456"

&#x20; }'



$userToken = $userResponse.accessToken



$loan = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Headers @{ Authorization = "Bearer $userToken" } `

&#x20; -Body '{

&#x20;   "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"

&#x20; }'



$loan

```



\---



\### Return Loan



Requires authentication.



```http

POST /loans/:id/return

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/$($loan.id)/return" `

&#x20; -Method POST `

&#x20; -Headers @{ Authorization = "Bearer $userToken" }

```



Response:



```json

{

&#x20; "id": "loan-uuid",

&#x20; "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",

&#x20; "status": "RETURNED",

&#x20; "loanDate": "2026-06-17T02:08:12.358975Z",

&#x20; "returnDate": "2026-06-17T02:20:00.000000Z",

&#x20; "createdAt": "2026-06-17T02:08:12.358975Z",

&#x20; "updatedAt": "2026-06-17T02:20:00.000000Z"

}

```



\---



\### List My Active Loans



Requires authentication.



```http

GET /loans/me/active

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/me/active" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $userToken" }

```



\---



\### List Loan History



Requires role: `ADMIN`.



```http

GET /loans/history

```



PowerShell:



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/history" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" }

```



If a normal user tries to access this endpoint:



```json

{

&#x20; "message": "Insufficient permissions",

&#x20; "error": "Forbidden",

&#x20; "statusCode": 403

}

```



\---



\## Internal Library Endpoints



These endpoints are used internally by `loans-service`.



They require:



```http

x-internal-api-key: local\_internal\_key

```



\### Get Internal Book



```http

GET /internal/books/:id

```



Response:



```json

{

&#x20; "id": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",

&#x20; "title": "Clean Code",

&#x20; "author": "Robert C. Martin",

&#x20; "isbn": "9780132350884",

&#x20; "year": 2008,

&#x20; "genre": "Software Engineering",

&#x20; "totalCopies": 5,

&#x20; "availableCopies": 5,

&#x20; "isAvailable": true

}

```



\---



\### Reserve Book Copy



```http

POST /internal/books/:id/reserve

```



This decreases `availableCopies` by 1 if there are available copies.



\---



\### Release Book Copy



```http

POST /internal/books/:id/release

```



This increases `availableCopies` by 1 if the book has borrowed copies.



\---



\## Direct Loans Service Endpoints



The `loans-service` exposes the following endpoints internally.



Base URL:



```text

http://localhost:8080

```



\### Health



```http

GET /health

```



Response:



```json

{

&#x20; "status": "ok",

&#x20; "service": "loans-service",

&#x20; "database": "ok"

}

```



\---



\### Register Loan



```http

POST /loans

```



Request:



```json

{

&#x20; "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"

}

```



Response:



```json

{

&#x20; "id": "loan-uuid",

&#x20; "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",

&#x20; "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",

&#x20; "status": "ACTIVE"

}

```



\---



\### Return Loan



```http

POST /loans/{id}/return

```



\---



\### List Active Loans by User



```http

GET /loans/users/{userId}/active

```



\---



\### List History



```http

GET /loans/history

```



\---



\# Main Flow



\## Full manual test with PowerShell



\### 1. Login as normal user



```powershell

$userResponse = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/auth/login" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Body '{

&#x20;   "email": "user@test.com",

&#x20;   "password": "123456"

&#x20; }'



$userToken = $userResponse.accessToken

```



\### 2. List books



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/books" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $userToken" }

```



\### 3. Create a loan



```powershell

$loan = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Headers @{ Authorization = "Bearer $userToken" } `

&#x20; -Body '{

&#x20;   "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"

&#x20; }'



$loan

```



\### 4. List active loans



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/me/active" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $userToken" }

```



\### 5. Return the loan



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/$($loan.id)/return" `

&#x20; -Method POST `

&#x20; -Headers @{ Authorization = "Bearer $userToken" }

```



\### 6. Login as admin



```powershell

$adminResponse = Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/auth/login" `

&#x20; -Method POST `

&#x20; -ContentType "application/json" `

&#x20; -Body '{

&#x20;   "email": "admin@test.com",

&#x20;   "password": "123456"

&#x20; }'



$adminToken = $adminResponse.accessToken

```



\### 7. List history as admin



```powershell

Invoke-RestMethod `

&#x20; -Uri "http://localhost:3000/loans/history" `

&#x20; -Method GET `

&#x20; -Headers @{ Authorization = "Bearer $adminToken" }

```



\---



\# Status Codes



Common status codes used by the API:



| Status | Meaning                                      |

| -----: | -------------------------------------------- |

|  `200` | Successful request                           |

|  `201` | Resource created                             |

|  `400` | Invalid request                              |

|  `401` | Missing or invalid authentication            |

|  `403` | Insufficient permissions                     |

|  `404` | Resource not found                           |

|  `409` | Conflict, for example duplicated active loan |

|  `502` | Upstream service communication error         |

|  `500` | Unexpected server error                      |



\---



\# Tests



\## Service A: library-service



Tests are implemented with Jest.



Current unit tests cover:



\* Creating a book with `availableCopies` equal to `totalCopies`.

\* Listing books with filters and pagination.

\* Throwing `NotFoundException` when a book does not exist.

\* Reserving one available copy.

\* Throwing `BadRequestException` when there are no available copies.



Run tests:



```powershell

cd library-service

npm test -- books.service.spec.ts

```



Expected result:



```text

PASS src/books/books.service.spec.ts

Tests: 5 passed

```



\---



\## Service B: loans-service



Tests are implemented with Go testing package.



Current unit tests cover:



\* Registering a loan and reserving the book.

\* Preventing duplicated active loans.

\* Releasing the book as compensation when loan creation fails.

\* Returning a loan and releasing the book.



Run tests:



```powershell

cd loans-service

go test ./...

```



Expected result:



```text

ok      github.com/jdeleonchang/library-technical-test/loans-service/internal/loans

```



\---



\# Useful Commands



\## Start all services



```powershell

docker compose up --build

```



\## Stop all services



```powershell

docker compose down

```



\## Stop all services and remove volumes



Use this when you want to reset databases completely.



```powershell

docker compose down -v

```



\## Rebuild without cache



```powershell

docker compose build --no-cache

docker compose up

```



\## Run only databases



```powershell

docker compose up -d library-db loans-db

```



\## Run library-service locally



```powershell

cd library-service

npm install

npx prisma migrate dev

npm run seed

npm run start:dev

```



\## Run loans-service locally



```powershell

cd loans-service

go mod tidy

go run ./cmd/server

```



\---



\# Troubleshooting



\## Prisma cannot reach database at `::1:5433`



Use `127.0.0.1` instead of `localhost` when running locally:



```powershell

$env:DATABASE\_URL="postgresql://postgres:postgres@127.0.0.1:5433/library\_db?schema=public"

npm run seed

```



\## Prisma generate fails during Docker build



The Dockerfile defines `DATABASE\_URL` during build so Prisma can generate the client.



\## `Cannot find module /app/dist/main`



The NestJS build output is under:



```text

dist/src/main.js

```



Therefore the Dockerfile starts the app with:



```text

node dist/src/main.js

```



\## Go service cannot find migrations



The Go Dockerfile copies migrations into the final image:



```dockerfile

COPY --from=builder /app/migrations ./migrations

```



Without this, the Go service would fail during startup because it runs `migrations/001\_init.sql`.



\---



\# Trade-offs



\## No distributed transaction



The system uses two independent databases. A distributed transaction was not implemented.



Instead, a compensation approach is used:



1\. `loans-service` reserves a book through `library-service`.

2\. Then it creates the loan in `loans-db`.

3\. If loan creation fails, it calls `release` in `library-service`.



This keeps the design simpler and closer to microservice-style ownership.



\## Internal API key instead of service-to-service JWT



Internal endpoints are protected with `x-internal-api-key`.



For a production system, this could be replaced or complemented by:



\* mTLS,

\* internal JWT,

\* API gateway policies,

\* private networking,

\* rotated secrets.



\## Basic seed data



The seed is designed to make evaluation easier. It creates stable users and a stable demo book.



\## Direct access to loans-service



`loans-service` exposes endpoints on port `8080` for testing and verification. In a production setup, it could be private and only accessible from `library-service`.



\---





\# Author



José De León





