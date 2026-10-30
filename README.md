# Chirpy
Boot.dev project building a simple server & RESTful api using Go and Postgres

This application acts as a super lightweight social platform:

- Register users and authenticate them
- Support for JWT login, along with refreshing and revoking tokens
- Create chirps (think simple tweets), view all chirps for a user, or delete chirps
- Support for health checks and metrics
- Simple file-serving capabilities

Keep in mind this is a toy project implementing a very basic social platform.

## Prerequistes

You will need to have the following tools installed or setup on your local machine

- **Golang (Go)** 1.26 or newer - <https://go.dev/doc/install>
- **PostgresSQL** (Any latest version) - <https://www.postgresql.org/download/>
- A `.env` file at the root for the required variabled needed for the project listed

Setup your system user password and database user password

- **goose** (for running database migrations though you can use what works best for you) - you can install using:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

To run goose migrations use the follow commands

Using your own DB URL connection

rollback

```sh
goose postgres "postgres://<USERNAME>:<PASSWORD>@localhost:5432/chirpy" down
```

apply migrations

```sh
goose postgres "postgres://<USERNAME>:<PASSWORD>@localhost:5432/chirpy" up
```

- **sqlc** (if you plan to add to change SQL queries) - <https://sqlc.dev> 

you can install using:

```sh
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

The Go code this command generates will be located in the `internal/database/` directory. To use run the command

```sh
sqlc generate
```

If you modify anything in `sql/queries` be sure to run the command above to generate the needed Go code.

## Installation

Install the latest version of this repo

```sh
go install github.com/Lilrosco/gator@latest
```

## Environment File

Create a `.env` file in your root directory for the project with the following content:

```
DB_URL="postgres://[USERNAME]:[PASSWORD]@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="your-own-secret-key-that-you-generate"
POLKA_KEY="your-own-polka-key"
```

**Note** - POLKA_KEY is to simulate having an external source like Stripe sending webhook requests to this application. You can put whatever value you want here.

Be sure to add your DB's username and password

## Building and running server

Build and start the server with:

`go build -o chirpy && ./chirpy`

The port that the app listens is set to `8080` by default.

You can reach the following endpoints:

- Frontend files: `http://localhost:8080/app/assets/...`

For example

```
http://localhost:8080/app/assets/logo.png

will retrieve the logo.png file located in the assets directory in the project.
```

- API base: `http://localhost:8080/api/...`
- Admin base: `http://localhost:8080/admin/...`

API Endpoints

**Authentication**

*Login*

Supports - `POST`

Example

```code
POST /api/login

Request Body

{
  "email": "walt@breakingbad.com",
  "password": "123456"
}
```

*Note* - In the response to this request a JWT / access token will be return as well as a refresh token. The JWT will live for only an hour.

*Refresh*

Supports - `POST`

Example

```code
POST /api/refresh

Headers - Authorization: Bearer {refreshToken}
```

*Note* - You will get a new access token that will live for another hour.

*Revoke*

Supports - `POST`

Example

```code
POST /api/revoke

Headers - Authorization: Bearer {refreshToken}
```

*Note* - This will invalidate the refresh token. Effectively terminating the logged in user's session.

**Users**

Supports - `POST, PUT`

Examples

```code
POST /api/users

Request Body

{
  "email": "saul@bettercall.com",
  "password": "123456"
}
```

```code
PUT /api/users

Header - Authorization: Bearer {jwtToken}

Request Body

{
  "email": "walter@breakingbad.com",
  "password": "losPollosHermanos"
}
```

**Chirps**

Supports - `POST, GET, DELETE`

Examples

```code
POST /api/chirps

Header - Authorization: Bearer {jwtToken}

Request Body

{
  "body": "I did it for me. I liked it. I was good at it. And I was really... I was alive."
}
```

```code
GET /api/chirps/

Header - Authorization: Bearer {jwtToken}

Query Params

author_id

/api/chirps?author_id={author_id}

sort

/api/chirps?sort=asc
/api/chirps?sort=desc
```

```code
GET /api/chirps/{chirpID}

Header - Authorization: Bearer {jwtToken}
```

```code
DELETE /api/chirps/{chirpID}

Header - Authorization: Bearer {jwtToken}
```

*Note* - You can only delete the chirp if user attached to it is logged in
