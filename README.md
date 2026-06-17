# Prueba Técnica - Sistema de Biblioteca

Sistema backend para una biblioteca implementado con dos servicios independientes:

- **Servicio A - `library-service`**: API principal expuesta al cliente.
- **Servicio B - `loans-service`**: servicio especializado en préstamos, implementado en Go.

El sistema permite autenticar usuarios, administrar libros, administrar usuarios, registrar préstamos, devolver préstamos, listar préstamos activos y consultar el histórico de préstamos.

Los servicios se comunican por HTTP y cada servicio tiene su propia base de datos PostgreSQL.

---

## Tabla de Contenido

- [Arquitectura](#arquitectura)
- [Stack Tecnológico](#stack-tecnológico)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Decisiones Técnicas](#decisiones-técnicas)
- [Requisitos Cubiertos](#requisitos-cubiertos)
- [Variables de Entorno](#variables-de-entorno)
- [Ejecución con Docker](#ejecución-con-docker)
- [Datos Demo](#datos-demo)
- [Documentación de Endpoints](#documentación-de-endpoints)
- [Flujo Principal](#flujo-principal)
- [Pruebas](#pruebas)
- [Comandos Útiles](#comandos-útiles)
- [Solución de Problemas](#solución-de-problemas)
- [Trade-offs](#trade-offs)
- [Checklist de Entrega](#checklist-de-entrega)

---

## Arquitectura

La solución está compuesta por dos servicios backend y dos bases de datos PostgreSQL independientes.

```mermaid
flowchart LR
    Client["Cliente / Consumidor API"]

    subgraph ServiceA["Servicio A: library-service"]
        Nest["NestJS API"]
        Auth["Auth JWT + Roles"]
        Books["Módulo Books"]
        Users["Módulo Users"]
        Internal["API Interna de Libros"]
        LoansProxy["Módulo Loans / Cliente HTTP"]
    end

    subgraph DB1["library-db"]
        UsersTable["Tabla users"]
        BooksTable["Tabla books"]
    end

    subgraph ServiceB["Servicio B: loans-service"]
        GoAPI["Go Chi API"]
        LoanBusiness["Lógica de Préstamos"]
        LibraryClient["Cliente HTTP hacia Library Service"]
    end

    subgraph DB2["loans-db"]
        LoansTable["Tabla loans"]
    end

    Client --> Nest
    Nest --> Auth
    Nest --> Books
    Nest --> Users
    Nest --> LoansProxy

    Books --> DB1
    Users --> DB1

    LoansProxy -->|"HTTP"| GoAPI
    GoAPI --> LoanBusiness
    LoanBusiness --> DB2
    LoanBusiness --> LibraryClient
    LibraryClient -->|"HTTP API Interna"| Internal
    Internal --> DB1
```

---

## Comunicación entre Servicios

El cliente consume principalmente el servicio `library-service`.

Cuando se registra un préstamo, el flujo es el siguiente:

```mermaid
sequenceDiagram
    participant Client as Cliente
    participant Library as library-service
    participant Loans as loans-service
    participant DBLoans as loans-db
    participant DBLibrary as library-db

    Client->>Library: POST /loans con JWT
    Library->>Loans: POST /loans
    Loans->>Library: POST /internal/books/:id/reserve
    Library->>DBLibrary: Disminuye availableCopies
    Library-->>Loans: Libro reservado
    Loans->>DBLoans: Crea préstamo ACTIVE
    Loans-->>Library: Préstamo creado
    Library-->>Client: 201 Created
```

Cuando se devuelve un préstamo:

```mermaid
sequenceDiagram
    participant Client as Cliente
    participant Library as library-service
    participant Loans as loans-service
    participant DBLoans as loans-db
    participant DBLibrary as library-db

    Client->>Library: POST /loans/:id/return con JWT
    Library->>Loans: POST /loans/:id/return
    Loans->>DBLoans: Marca préstamo como RETURNED
    Loans->>Library: POST /internal/books/:id/release
    Library->>DBLibrary: Aumenta availableCopies
    Loans-->>Library: Préstamo devuelto
    Library-->>Client: 200 OK
```

---

## Stack Tecnológico

### Servicio A: `library-service`

- Node.js
- NestJS
- Prisma ORM
- PostgreSQL
- JWT
- Roles `ADMIN` / `USER`
- Class Validator
- Docker

### Servicio B: `loans-service`

- Go
- Chi Router
- PostgreSQL
- `database/sql`
- `pgx`
- Docker

### Infraestructura

- Docker Compose
- Dos bases de datos PostgreSQL
- Contenedores independientes por servicio

---

## Estructura del Proyecto

```text
library-technical-test/
  docker-compose.yml
  README.md
  .gitignore

  library-service/
    Dockerfile
    .dockerignore
    .env.example
    prisma/
      schema.prisma
      seed.cjs
      migrations/
    src/
      auth/
      books/
      database/
      internal/
      loans/
      users/

  loans-service/
    Dockerfile
    .dockerignore
    .env.example
    go.mod
    go.sum
    cmd/
      server/
        main.go
    migrations/
      001_init.sql
    internal/
      loans/
      platform/
        database/
        httpclient/
```

---

## Decisiones Técnicas

### 1. `library-service` como API principal

El cliente consume `library-service`, ya que este servicio centraliza:

- autenticación,
- autorización,
- usuarios,
- libros,
- entrada al flujo de préstamos.

Los endpoints de préstamos en `library-service` funcionan como una capa de entrada que llama internamente a `loans-service`.

### 2. Bases de datos independientes

Cada servicio tiene su propia base de datos:

- `library-db`: almacena usuarios y libros.
- `loans-db`: almacena préstamos.

Esto evita que un servicio acceda directamente a la base de datos del otro y mantiene una separación clara de responsabilidades.

### 3. Comunicación HTTP entre servicios

Los servicios se comunican mediante REST sobre HTTP.

- `library-service` llama a `loans-service` para registrar y devolver préstamos.
- `loans-service` llama a endpoints internos de `library-service` para reservar o liberar copias de libros.

### 4. Protección de endpoints internos

Los endpoints internos de `library-service` se protegen usando el header:

```http
x-internal-api-key: local_internal_key
```

Esto evita que un cliente externo reserve o libere copias directamente sin pasar por el flujo correcto de préstamos.

### 5. Compensación en lugar de transacciones distribuidas

Como existen dos bases de datos independientes, no se implementó una transacción distribuida.

En su lugar, se usa una estrategia de compensación:

1. `loans-service` reserva una copia del libro llamando a `library-service`.
2. Luego intenta crear el préstamo en `loans-db`.
3. Si falla la creación del préstamo, `loans-service` llama a `release` para devolver la copia reservada.

---

## Requisitos Cubiertos

- CRUD de libros.
- CRUD de usuarios.
- Autenticación JWT.
- Roles `ADMIN` y `USER`.
- Filtros y paginación de libros.
- Registro de préstamos.
- Devolución de préstamos.
- Listado de préstamos activos por usuario.
- Listado de histórico de préstamos.
- Comunicación HTTP entre servicios.
- Persistencia independiente por servicio.
- Docker Compose para levantar servicios y bases de datos.
- Seed automático para datos demo.
- Tests unitarios para ambos servicios.
- Documentación de arquitectura, endpoints, decisiones y flujo completo.

---

## Variables de Entorno

### `library-service/.env.example`

```env
PORT=3000
DATABASE_URL=postgresql://postgres:postgres@localhost:5433/library_db?schema=public
JWT_SECRET=local_super_secret_key
JWT_EXPIRES_IN=1d
LOANS_SERVICE_URL=http://localhost:8080
INTERNAL_API_KEY=local_internal_key
```

### `loans-service/.env.example`

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5434/loans_db?sslmode=disable
LIBRARY_SERVICE_URL=http://localhost:3000
INTERNAL_API_KEY=local_internal_key
```

Cuando se ejecuta con Docker Compose, estas variables ya están configuradas dentro de `docker-compose.yml`.

---

## Ejecución con Docker

### Requisitos previos

- Docker Desktop instalado.
- Docker Compose disponible.
- Puertos disponibles:
  - `3000`
  - `8080`
  - `5433`
  - `5434`

### Levantar todo el sistema

Desde la raíz del proyecto:

```powershell
docker compose up --build
```

Este comando levanta:

- `library-service`
- `loans-service`
- `library-db`
- `loans-db`

El contenedor de `library-service` ejecuta automáticamente:

```text
npx prisma migrate deploy
npm run seed
node dist/src/main.js
```

Esto significa que:

1. se aplican las migraciones,
2. se crean datos demo,
3. se levanta la API de NestJS.

### Validar servicios

En otra terminal:

```powershell
Invoke-RestMethod -Uri "http://localhost:3000" -Method GET
```

Respuesta esperada:

```text
Hello World!
```

Validar el servicio Go:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
```

Respuesta esperada:

```json
{
  "status": "ok",
  "service": "loans-service",
  "database": "ok"
}
```

---

## Datos Demo

El seed crea automáticamente los siguientes datos:

### Usuario Administrador

```text
Email: admin@test.com
Password: 123456
Role: ADMIN
```

### Usuario Normal

```text
Email: user@test.com
Password: 123456
Role: USER
```

### Libro Demo

```text
ID: 117f2aaf-61ab-4b49-b34a-7b331f6947a8
Title: Clean Code
Author: Robert C. Martin
ISBN: 9780132350884
Year: 2008
Genre: Software Engineering
Total copies: 5
Available copies: 5
```

---

# Documentación de Endpoints

URL base del servicio principal:

```text
http://localhost:3000
```

URL base del servicio de préstamos:

```text
http://localhost:8080
```

El cliente debe consumir principalmente `library-service`.

---

## Autenticación

### Login

```http
POST /auth/login
```

Request:

```json
{
  "email": "admin@test.com",
  "password": "123456"
}
```

Response:

```json
{
  "accessToken": "jwt_token_here",
  "user": {
    "id": "e11067e4-7927-442b-89ea-533cf5564609",
    "name": "Admin User",
    "email": "admin@test.com",
    "role": "ADMIN"
  }
}
```

Ejemplo en PowerShell:

```powershell
$adminResponse = Invoke-RestMethod `
  -Uri "http://localhost:3000/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{
    "email": "admin@test.com",
    "password": "123456"
  }'

$adminToken = $adminResponse.accessToken
```

---

## Endpoints de Libros

### Crear Libro

Requiere rol: `ADMIN`

```http
POST /books
```

Request:

```json
{
  "title": "Clean Architecture",
  "author": "Robert C. Martin",
  "isbn": "9780134494166",
  "year": 2017,
  "genre": "Software Engineering",
  "totalCopies": 3
}
```

Response:

```json
{
  "id": "book-uuid",
  "title": "Clean Architecture",
  "author": "Robert C. Martin",
  "isbn": "9780134494166",
  "year": 2017,
  "genre": "Software Engineering",
  "totalCopies": 3,
  "availableCopies": 3,
  "createdAt": "2026-06-17T00:00:00.000Z",
  "updatedAt": "2026-06-17T00:00:00.000Z"
}
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/books" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $adminToken" } `
  -Body '{
    "title": "Clean Architecture",
    "author": "Robert C. Martin",
    "isbn": "9780134494166",
    "year": 2017,
    "genre": "Software Engineering",
    "totalCopies": 3
  }'
```

### Listar Libros

Requiere autenticación.

```http
GET /books
```

Query params disponibles:

| Nombre | Tipo | Requerido | Descripción |
|---|---:|---:|---|
| `author` | string | No | Filtra por autor |
| `genre` | string | No | Filtra por género |
| `available` | boolean | No | Filtra libros disponibles o no disponibles |
| `page` | number | No | Número de página |
| `limit` | number | No | Cantidad de registros por página |

Ejemplo:

```http
GET /books?author=Robert&available=true&page=1&limit=10
```

Response:

```json
{
  "items": [
    {
      "id": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",
      "title": "Clean Code",
      "author": "Robert C. Martin",
      "isbn": "9780132350884",
      "year": 2008,
      "genre": "Software Engineering",
      "totalCopies": 5,
      "availableCopies": 5
    }
  ],
  "meta": {
    "total": 1,
    "page": 1,
    "limit": 10,
    "totalPages": 1
  }
}
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/books" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $adminToken" }
```

### Obtener Libro por ID

Requiere autenticación.

```http
GET /books/:id
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/books/117f2aaf-61ab-4b49-b34a-7b331f6947a8" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $adminToken" }
```

### Actualizar Libro

Requiere rol: `ADMIN`

```http
PATCH /books/:id
```

Request:

```json
{
  "title": "Updated title",
  "totalCopies": 10
}
```

### Eliminar Libro

Requiere rol: `ADMIN`

```http
DELETE /books/:id
```

Response:

```json
{
  "message": "Book deleted successfully"
}
```

---

## Endpoints de Usuarios

Todos los endpoints de usuarios requieren rol `ADMIN`.

### Crear Usuario

```http
POST /users
```

Request:

```json
{
  "name": "Normal User",
  "email": "user@test.com",
  "password": "123456",
  "role": "USER"
}
```

Response:

```json
{
  "id": "user-uuid",
  "name": "Normal User",
  "email": "user@test.com",
  "role": "USER",
  "createdAt": "2026-06-17T00:00:00.000Z",
  "updatedAt": "2026-06-17T00:00:00.000Z"
}
```

La contraseña nunca se retorna en la respuesta.

### Listar Usuarios

```http
GET /users
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/users" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $adminToken" }
```

### Obtener Usuario por ID

```http
GET /users/:id
```

### Actualizar Usuario

```http
PATCH /users/:id
```

Request:

```json
{
  "name": "Updated User",
  "role": "ADMIN"
}
```

### Eliminar Usuario

```http
DELETE /users/:id
```

Response:

```json
{
  "message": "User deleted successfully"
}
```

---

## Endpoints de Préstamos desde `library-service`

Estos son los endpoints que debería consumir el cliente.

### Registrar Préstamo

Requiere autenticación.

```http
POST /loans
```

Request como usuario normal:

```json
{
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"
}
```

Request como admin para otro usuario:

```json
{
  "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"
}
```

Response:

```json
{
  "id": "loan-uuid",
  "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",
  "status": "ACTIVE",
  "loanDate": "2026-06-17T02:08:12.358975Z",
  "createdAt": "2026-06-17T02:08:12.358975Z",
  "updatedAt": "2026-06-17T02:08:12.358975Z"
}
```

Ejemplo:

```powershell
$userResponse = Invoke-RestMethod `
  -Uri "http://localhost:3000/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{
    "email": "user@test.com",
    "password": "123456"
  }'

$userToken = $userResponse.accessToken

$loan = Invoke-RestMethod `
  -Uri "http://localhost:3000/loans" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $userToken" } `
  -Body '{
    "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"
  }'

$loan
```

### Devolver Préstamo

Requiere autenticación.

```http
POST /loans/:id/return
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/$($loan.id)/return" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $userToken" }
```

Response:

```json
{
  "id": "loan-uuid",
  "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",
  "status": "RETURNED",
  "loanDate": "2026-06-17T02:08:12.358975Z",
  "returnDate": "2026-06-17T02:20:00.000000Z",
  "createdAt": "2026-06-17T02:08:12.358975Z",
  "updatedAt": "2026-06-17T02:20:00.000000Z"
}
```

### Listar Mis Préstamos Activos

Requiere autenticación.

```http
GET /loans/me/active
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/me/active" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $userToken" }
```

### Listar Histórico de Préstamos

Requiere rol: `ADMIN`.

```http
GET /loans/history
```

Ejemplo:

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/history" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $adminToken" }
```

Si un usuario normal intenta acceder:

```json
{
  "message": "Insufficient permissions",
  "error": "Forbidden",
  "statusCode": 403
}
```

---

## Endpoints Internos de `library-service`

Estos endpoints son utilizados internamente por `loans-service`.

Requieren el header:

```http
x-internal-api-key: local_internal_key
```

### Obtener Libro Interno

```http
GET /internal/books/:id
```

Response:

```json
{
  "id": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",
  "title": "Clean Code",
  "author": "Robert C. Martin",
  "isbn": "9780132350884",
  "year": 2008,
  "genre": "Software Engineering",
  "totalCopies": 5,
  "availableCopies": 5,
  "isAvailable": true
}
```

### Reservar Copia de Libro

```http
POST /internal/books/:id/reserve
```

Disminuye `availableCopies` en 1 si hay copias disponibles.

### Liberar Copia de Libro

```http
POST /internal/books/:id/release
```

Aumenta `availableCopies` en 1 si existen copias prestadas.

---

## Endpoints Directos de `loans-service`

`loans-service` expone endpoints en el puerto `8080`.

Base URL:

```text
http://localhost:8080
```

### Health

```http
GET /health
```

Response:

```json
{
  "status": "ok",
  "service": "loans-service",
  "database": "ok"
}
```

### Registrar Préstamo

```http
POST /loans
```

Request:

```json
{
  "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"
}
```

Response:

```json
{
  "id": "loan-uuid",
  "userId": "81b8f134-ba1f-42cb-a834-859b9ac23f09",
  "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8",
  "status": "ACTIVE"
}
```

### Devolver Préstamo

```http
POST /loans/{id}/return
```

### Listar Préstamos Activos por Usuario

```http
GET /loans/users/{userId}/active
```

### Listar Histórico

```http
GET /loans/history
```

---

# Flujo Principal

## Prueba manual completa con PowerShell

### 1. Login como usuario normal

```powershell
$userResponse = Invoke-RestMethod `
  -Uri "http://localhost:3000/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{
    "email": "user@test.com",
    "password": "123456"
  }'

$userToken = $userResponse.accessToken
```

### 2. Listar libros

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/books" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $userToken" }
```

### 3. Crear préstamo

```powershell
$loan = Invoke-RestMethod `
  -Uri "http://localhost:3000/loans" `
  -Method POST `
  -ContentType "application/json" `
  -Headers @{ Authorization = "Bearer $userToken" } `
  -Body '{
    "bookId": "117f2aaf-61ab-4b49-b34a-7b331f6947a8"
  }'

$loan
```

### 4. Listar préstamos activos

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/me/active" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $userToken" }
```

### 5. Devolver préstamo

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/$($loan.id)/return" `
  -Method POST `
  -Headers @{ Authorization = "Bearer $userToken" }
```

### 6. Login como admin

```powershell
$adminResponse = Invoke-RestMethod `
  -Uri "http://localhost:3000/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{
    "email": "admin@test.com",
    "password": "123456"
  }'

$adminToken = $adminResponse.accessToken
```

### 7. Consultar histórico como admin

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:3000/loans/history" `
  -Method GET `
  -Headers @{ Authorization = "Bearer $adminToken" }
```

---

# Status Codes

| Status | Significado |
|---:|---|
| `200` | Solicitud exitosa |
| `201` | Recurso creado |
| `400` | Solicitud inválida |
| `401` | Autenticación ausente o inválida |
| `403` | Permisos insuficientes |
| `404` | Recurso no encontrado |
| `409` | Conflicto, por ejemplo préstamo activo duplicado |
| `502` | Error de comunicación con otro servicio |
| `500` | Error inesperado del servidor |

---

# Pruebas

## Servicio A: `library-service`

Las pruebas están implementadas con Jest.

Actualmente cubren:

- Crear un libro con `availableCopies` igual a `totalCopies`.
- Listar libros con filtros y paginación.
- Lanzar `NotFoundException` cuando el libro no existe.
- Reservar una copia disponible.
- Lanzar `BadRequestException` cuando no hay copias disponibles.

Ejecutar pruebas:

```powershell
cd library-service
npm test -- books.service.spec.ts
```

Resultado esperado:

```text
PASS src/books/books.service.spec.ts
Tests: 5 passed
```

## Servicio B: `loans-service`

Las pruebas están implementadas con el paquete estándar de testing de Go.

Actualmente cubren:

- Registrar un préstamo y reservar el libro.
- Evitar préstamos activos duplicados.
- Liberar el libro como compensación cuando falla la creación del préstamo.
- Devolver un préstamo y liberar el libro.

Ejecutar pruebas:

```powershell
cd loans-service
go test ./...
```

Resultado esperado:

```text
ok      github.com/jdeleonchang/library-technical-test/loans-service/internal/loans
```

---

# Comandos Útiles

## Levantar todos los servicios

```powershell
docker compose up --build
```

## Detener todos los servicios

```powershell
docker compose down
```

## Detener servicios y eliminar volúmenes

Usar cuando se quiera reiniciar completamente las bases de datos.

```powershell
docker compose down -v
```

## Rebuild sin caché

```powershell
docker compose build --no-cache
docker compose up
```

## Levantar solo las bases de datos

```powershell
docker compose up -d library-db loans-db
```

## Ejecutar `library-service` localmente

```powershell
cd library-service
npm install
npx prisma migrate dev
npm run seed
npm run start:dev
```

## Ejecutar `loans-service` localmente

```powershell
cd loans-service
go mod tidy
go run ./cmd/server
```

---

# Solución de Problemas

## Prisma no puede conectarse a `::1:5433`

Cuando se ejecuta localmente en Windows, puede ocurrir que Prisma intente conectarse usando IPv6.

Solución:

```powershell
$env:DATABASE_URL="postgresql://postgres:postgres@127.0.0.1:5433/library_db?schema=public"
npm run seed
```

## Prisma generate falla durante Docker build

El Dockerfile define `DATABASE_URL` durante el build para que Prisma pueda generar el cliente correctamente.

## Error `Cannot find module /app/dist/main`

El build de NestJS genera la salida en:

```text
dist/src/main.js
```

Por eso el Dockerfile inicia la aplicación con:

```text
node dist/src/main.js
```

## Go no encuentra las migraciones

El Dockerfile de Go copia las migraciones a la imagen final:

```dockerfile
COPY --from=builder /app/migrations ./migrations
```

Sin esto, el servicio Go fallaría al iniciar porque ejecuta `migrations/001_init.sql`.

---

# Trade-offs

## No se implementó transacción distribuida

El sistema usa dos bases de datos independientes.

En lugar de una transacción distribuida, se usa una estrategia de compensación:

1. `loans-service` reserva el libro por medio de `library-service`.
2. Luego crea el préstamo en `loans-db`.
3. Si la creación del préstamo falla, llama a `release` en `library-service`.

Esto mantiene el diseño más simple y cercano a un enfoque de microservicios.

## API key interna en lugar de JWT entre servicios

Los endpoints internos usan `x-internal-api-key`.

En un sistema productivo podría complementarse o reemplazarse por:

- mTLS,
- JWT interno entre servicios,
- API Gateway,
- redes privadas,
- rotación de secretos.

## Seed de datos demo

El seed existe para facilitar la evaluación del proyecto.

Crea usuarios y un libro base para que el evaluador pueda probar el sistema inmediatamente después de ejecutar Docker Compose.

## Acceso directo a `loans-service`

`loans-service` expone endpoints en el puerto `8080` para pruebas y validación.

En un ambiente productivo, este servicio podría ser privado y accesible únicamente desde `library-service`.

---


---

# Autor

José De León



