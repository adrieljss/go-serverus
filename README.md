<div style="text-align:center;display:grid;align-items:center;justify-items:center;margin-bottom:2.5rem">
<img src="./public/logo.png" alt="GoServerus logo" style="width:30rem" />
<p style="margin:25px 0 2px 0;font-size:2rem;font-weight:bold">GoServerus</p>
<p style="font-size:1rem;width:25rem;">A flexible and up-to-standard boilerplate code for golang backends intended for mobile applications.</p>
</div>

# Features
| Feature                                  | Status |
|------------------------------------------|--------|
| [Basic Authentication](#authentication)  | ✅     |
| [PostgreSQL Database Handler](#database) | ✅     |
| [Redis Caching](#caching)                | ✅     |
| [Ratelimiter](#ratelimiter)              | ✅     |
| Categorized Directories                  | ✅     |
| Easy Error Handling                      | ✅     |
| Log System                               | ✅     |
| [Documentations](#documentations)        | TODO   |

# Getting Started
Before starting, please read [Database](#database) and [Caching](#caching) to know how the database and caching systems are supposed to be handled. 
## Development
For development purposes, it is recommended to start a seperate redis and PostgreSQL server in localhost, (and set up the .env files) and then run with:
```
$ go run .
```

## Production
For production, use docker-compose, so that the program is dockerized and redis starts in the same server as the API.

Start docker-compose with:
```
$ docker compose up
```

# Authentication
GoServerus uses a JWT basic authentication without refresh tokens. 

> [!NOTE]
> Refresh tokens are not used because it is intended for users (in their mobile apps) to store username and password, and that will be used to "refresh" the token.

> [!WARNING]
> If you think this is not secure, it is recommended to tweak the authentication system.

Authentication features include:
- Email confirmations (can use any SMTP server)
- Basic Login/Register
- OAuth2 Sign In
- Forgot Password

# Database
GoServerus uses PostgreSQL for its database. Configurations are recommended to be included in the .env file.

It is intended that the database server is seperate from the API server, [Caching](#caching) is used to speed up requests.

Current .env configurations:
```
POSTGRES_URI = "postgres://postgres:password@localhost:5432/serverus_db"
```
please customize as needed.

# Caching
Caching uses redis that is in the same server as the API, it is **NOT** recommended to use redis seperate from the API server, because the network latency will nullify caching speed-up effects.

To make it easier, there is a `docker-compose.yml` file that has been customized to dockerize the API, and also start redis in the same server.

# Ratelimiter
A ratelimiter is used so that a user cannot spam the API (and take down the server, such as DDOS attacks).

Its configurations can be changed in the `env/env.go` file, **NOT** in the .env file

# Directories
To make GoServerus as flexible and as friendly to developers as possible

# Documentations
Currently, documentations are not available. Some functions might have a bit of documentations, but they are currently not formatted neatly.

Important or dangerous functions always have notes of when to use them and how.