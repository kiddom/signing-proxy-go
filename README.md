Signing Proxy
=====
Proxy to sign kiddom api requests.

Usage
----

The proxy reads the following environment variables:
- `REMOTE_HOST` is the Kiddom content api in the form of `https://kiddom-content-dev.herokuapp.com`.
- `PORT` is the local port that the proxy will listen on.
- `K_USER_ID` is the user id.
- `K_PRIVATE_KEY` is the private key used to sign the requests.

### Example:
```
$ REMOTE_HOST={api_host} PORT={local_port} K_USER_ID={user_id} K_PRIVATE_KEY={private_key/password} \
    go run main.go
```
or set up a `.env` file that will be read by the proxy
```
$ cat <<EOF >> .env
REMOTE_HOST="https://kiddom-content-dev.herokuapp.com"
PORT=4000
K_USER_ID=foobar
K_PRIVATE_KEY=myvoiceismypassword
EOF
$ go run main.go
```

Then you can pass requests directly to the proxy, which will sign them:
```
$ curl -sH 'Accept-encoding: identity' http://localhost:4000/content_import
```

Dependencies
----

The proxy uses [github.com/joho/godotenv](https://github.com/joho/godotenv) to read environment variables
from a `.env` file.
