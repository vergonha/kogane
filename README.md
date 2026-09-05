<h1 align="center">
  kogane ‎ . ݁⋆ ۶ৎ ݁˖ . ݁
</h1>
<p align="center">a private manga library, for one reader only</p>

<img src="https://i.imgur.com/JyrKAc8.png" style="align-self: flex-start;" />


---

## 𐔌՞. .՞𐦯    200gb sitting on a hard drive, doing nothing

at some point i had downloaded around 200gb of manga and just left it there on my hard drive. that's it, that was the whole system. i could only read it on the one machine it was sitting on, which defeats the entire point of having a library.

so i went shopping for somewhere to put it. tried a few of the usual suspects people recommend for this kind of thing and ended up on cloudflare r2, mostly because it's s3 compatible and doesn't charge you for pulling your own files out, which every other provider seems to think is a reasonable thing to charge for.

kogane is what came out of that. it's my server, reading my manga, and nobody else is getting an account. this isn't a product, it's a fix for a very specific problem i had.

<img src="https://i.imgur.com/xvIjPEy.png" style="align-self: flex-start;" />

---

### ᨒ⋆.ೃ࿔ security ──

i didn't want to ship something that falls over the first time someone pokes at it, even if the audience is exactly one person.

already in place:
- [x] bcrypt for passwords, plus a dummy hash check on unknown usernames so login timing doesn't leak whether an account exists
- [x] every title/vol from a query string goes through `library.ValidComponent` before touching an r2 key, avoiding path traversal
- [x] turnstile in front of the login form so bots don't get a free swing at credential stuffing
- [x] app never streams files itself, pdfs and covers only go out as short-lived presigned urls
- [x] templates go through `html/template`, auto-escaped, so mangadex metadata can't turn into markup
- [x] session and csrf tokens compared with `crypto/subtle.ConstantTimeCompare`
- [x] logging in kills every other session that user had open
- [x] expired sessions rejected on lookup and swept out every 15 minutes
- [x] cookies set `HttpOnly`, `SameSite=Lax`, and `Secure` outside dev mode
- [x] parameterized queries everywhere

still missing, fine for now but wouldn't be if this stopped being a one-person server:
- [ ] dependency CVE scanning in CI
- [ ] 2fa
- [ ] rate limiting on login, turnstile alone won't stop a patient human doing it by hand
- [ ] login audit trail
- [ ] security headers, CSP, HSTS, would need to live here or in whatever sits in front of it
- [ ] real password validation, right now it's just "not empty"
- [ ] timeouts on `http.ListenAndServe`, it's running go's defaults which is to say none
- [ ] fix `Turnstile.Verify` calling `log.Fatal` on a missing secret key, kills the whole process on the first login instead of failing just that request, this one's an actual bug

---

### ⊹ ࣪ ˖ ໒꒱ implementation ──

| Layer            | Choice                                            |
| ---------------- | -------------------------------------------------- |
| backend          | go, stdlib `net/http`, no framework                 |
| routing          | plain `http.ServeMux`                               |
| database         | sqlite via `modernc.org/sqlite`, no cgo             |
| file storage     | manga volumes as pdfs, sitting in an r2 bucket       |
| delivery         | short-lived presigned urls, app never serves the files directly |
| auth             | cookie sessions, csrf tokens, bcrypt                |
| bot gate         | cloudflare turnstile                               |
| metadata         | `library.json`, occasionally topped up against the mangadex api |
| front-end        | plain html/css/js, no build step                    |
| deploy           | github actions builds a binary and rsyncs it to a vps under systemd |

the folder layout mirrors how `cmd/server/main.go` wires things up. `config` reads the env vars, `database` opens sqlite and hands out the repositories, `library` loads and normalizes `library.json`, `storage` wraps the r2 presigning client, `auth` and `turnstile` handle sessions/csrf/captcha, and `http/handlers` plus `http/router.go` glue it all together behind the auth middleware. `cmd/cli` is its own small binary just for adding users once an admin already exists.

---

### ⋆ᨳଓ routes ──

| Method   | Endpoint                                | What it does                                |
| -------- | ---------------------------------------- | -------------------------------------------- |
| `GET`    | `/`                                       | login, or the admin registration screen on first run |
| `POST`   | `/`                                       | login submit                                |
| `POST`   | `/logout`                                 | ends the session                            |
| `GET`    | `/dashboard`                              | the library grid                            |
| `GET`    | `/manga`                                  | volume list for one title                   |
| `GET`    | `/read`                                   | the reader                                  |
| `GET`    | `/pdf`                                    | redirects to a presigned r2 url             |
| `GET`    | `/cover`                                  | cover art                                   |
| `GET`    | `/api/progress`                           | reading progress, everything                |
| `GET`    | `/api/progress/{mangadex_id}`             | reading progress for one title              |
| `POST`   | `/api/progress`                           | save progress                               |
| `POST`   | `/api/progress/{mangadex_id}/complete`    | mark a title as finished                    |
| `DELETE` | `/api/progress/{mangadex_id}`             | wipe stored progress                        |

anything past `/` needs a session, and the api routes also need a valid csrf token.

---

### ⋆˚꩜｡ּ running it ──

you need an r2 bucket and a turnstile site set up before this does anything useful. fill these into `.env` (there's an `.env.example` for reference):

```
KOGANE_DEVELOPMENT=
KOGANE_SERVER_PORT=
CLOUDFLARE_TURNSTILE_SECRET_KEY=
CLOUDFLARE_TURNSTILE_SITE_KEY=
R2_BUCKET_NAME=
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
KOGANE_LIBRARY_PATH=
KOGANE_TEMPLATES_GLOB=
```

⤿ clone and run
```
git clone <repo-url>
cd kogane
cp .env.example .env
go run ./cmd/server
```

first time you hit `/` with an empty database, it drops you into registration instead of login. whoever signs up there becomes the admin, and there's only ever one.

⤿ adding users after that
```
go run ./cmd/cli create-user <username> <password>
```

⤿ building for deploy
```
go build -trimpath -ldflags="-s -w" -o build/server ./cmd/server
```

the actual deploys just happen on push to `main`, see `.github/workflows/deploy.yml`. it cross-compiles for linux/amd64, rsyncs the binary plus `static/` and `templates/` to the vps, and bounces the `manga` systemd unit.

<img src="https://i.imgur.com/ZwXbykb.png" style="align-self: flex-start;" />


---

### ᯓ★ what to test ──

log in, click into a title, open a volume, flip through a few pages, close it. go back to the dashboard and it should show where you left off. log out and back in and that progress should still be there, since it's sitting in sqlite and never touched the browser.

---

### ˖ ִֶ♱ roadmap ──

- [x] session auth with bcrypt and csrf
- [x] turnstile on login, first-run admin setup
- [x] r2-backed storage with presigned delivery
- [x] per-user reading progress
- [x] manga detail page and the library grid redesign
- [x] moved off `mattn/go-sqlite3` (cgo) onto `modernc.org/sqlite`
- [ ] some kind of rate limit on login, turnstile alone isn't enough forever
- [ ] better handling for `library.json` entries missing mangadex metadata
- [ ] rewriting the css, most of it is inherited from a 2023 version of this idea and it shows

---

### notes ──

this isn't multi-tenant and it's not going to be. one admin manages the user list, that's the whole model. `library.json` is the source of truth for the catalog, so anything listed there needs a matching folder in the bucket or it just won't resolve. reading progress lives entirely in sqlite, separate from the library metadata, so wiping one doesn't touch the other.
