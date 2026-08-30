package main

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbPath             = "./manga.db"
	sessionName        = "session_id"
	bcryptCost         = 12
	turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
)

type turnstileResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

var (
	db            *sql.DB
	tmpl          *template.Template
	isDevelopment bool
	dummyHash     []byte
	dataDir       string
	addr          string
)

type Manga struct {
	Title       string
	Volumes     []string
	Cover       string
	Description string
}

func verifyTurnstile(r *http.Request) bool {
	secret := os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY")
	if secret == "" {
		log.Println("CLOUDFLARE_TURNSTILE_SECRET_KEY não configurada")
		return false
	}

	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		return false
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		form.Set("remoteip", ip)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		turnstileVerifyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return false
	}

	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Turnstile: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result turnstileResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Success
}

func initDB() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id       INTEGER PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			hash     TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			csrf_token TEXT NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	pass := os.Getenv("KOGANE_ADMIN_PASS")
	if pass == "" {
		log.Fatal("KOGANE_ADMIN_PASS não configurada")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcryptCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`INSERT OR IGNORE INTO users (username, hash) VALUES ('admin', ?)`,
		string(hash),
	)
	return err
}

func getSessionUserID(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return 0, false
	}
	var userID int64
	var expires int64
	err = db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE id = ?`,
		cookie.Value,
	).Scan(&userID, &expires)
	if err != nil || time.Now().Unix() > expires {
		return 0, false
	}
	return userID, true
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getSessionUserID(r); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func deleteSession(r *http.Request) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return
	}
	db.Exec(`DELETE FROM sessions WHERE id = ?`, cookie.Value)
}

func renderTemplate(w http.ResponseWriter, name string, data any) error {
	if isDevelopment {
		t, err := template.ParseGlob("templates/*.html")
		if err != nil {
			return err
		}

		return t.ExecuteTemplate(w, name, data)
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	var err error

	isDevelopment = os.Getenv("KOGANE_DEVELOPMENT") == "true"

	addr = os.Getenv("KOGANE_SERVER_PORT")
	if addr == "" {
		addr = ":8080"
	}

	dataDir = os.Getenv("KOGANE_DATA_PATH")
	if dataDir == "" {
		dataDir = "E:\\"
	}

	db, err = sql.Open("sqlite3", "file:manga.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatal(err)
	}

	if db == nil {
		log.Fatal("db is nil")
	}

	if err = initDB(); err != nil {
		log.Fatal(err)
	}

	dummyHash, err = bcrypt.GenerateFromPassword([]byte("dummy-kogane"), bcryptCost)
	if err != nil {
		log.Fatal(err)
	}

	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /login", handleLoginPage)
	mux.HandleFunc("POST /login", handleLoginSubmit)
	mux.HandleFunc("POST /logout", requireAuth(handleLogout))
	mux.HandleFunc("GET /", requireAuth(handleDashboard))
	mux.HandleFunc("GET /read", requireAuth(handleReader))
	mux.HandleFunc("GET /pdf", requireAuth(handlePDF))
	mux.HandleFunc("GET /cover", requireAuth(handleCover))

	log.Printf("Servidor em %s", addr)
	log.Printf("Development mode: %v", isDevelopment)

	log.Fatal(http.ListenAndServe(addr, mux))
}

func handlePDF(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if title == "" || vol == "" {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	if strings.Contains(title, "..") ||
		strings.Contains(vol, "..") ||
		strings.ContainsAny(title, "/\\") ||
		strings.ContainsAny(vol, "/\\") {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	dataDirAbs, err := filepath.Abs(dataDir)
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	target := filepath.Join(dataDirAbs, title, vol)

	// Validação do caminho final.
	rel, err := filepath.Rel(dataDirAbs, target)
	if err != nil {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	if rel == ".." ||
		strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, target)
}

func handleReader(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !validLibraryComponent(title) ||
		!validLibraryComponent(vol) ||
		!strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	renderTemplate(w, "reader.html", map[string]string{
		"Title": title,
		"Vol":   vol,
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	mangas, err := scanLibrary(dataDir)
	if err != nil {
		http.Error(w, "Erro ao ler biblioteca", http.StatusInternalServerError)
		return
	}

	csrfToken, ok := getCSRFToken(r)
	if !ok {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	renderTemplate(w, "dashboard.html", map[string]any{
		"Mangas":    mangas,
		"CSRFToken": csrfToken,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if !requireCSRF(r) {
		http.Error(w, "CSRF inválido", http.StatusForbidden)
		return
	}

	deleteSession(r)
	http.SetCookie(w, &http.Cookie{
		Name:   sessionName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func newSession(userID int64) (string, string, error) {
	sessionBytes := make([]byte, 32)
	if _, err := rand.Read(sessionBytes); err != nil {
		return "", "", err
	}

	csrfBytes := make([]byte, 32)
	if _, err := rand.Read(csrfBytes); err != nil {
		return "", "", err
	}

	sessionID := hex.EncodeToString(sessionBytes)
	csrfToken := hex.EncodeToString(csrfBytes)
	expires := time.Now().Add(2 * time.Hour).Unix()

	_, err := db.Exec(`
		INSERT INTO sessions
			(id, user_id, expires_at, csrf_token)
		VALUES (?, ?, ?, ?)
	`, sessionID, userID, expires, csrfToken)

	if err != nil {
		return "", "", err
	}

	return sessionID, csrfToken, nil
}

func getCSRFToken(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return "", false
	}

	var token string
	var expires int64

	err = db.QueryRow(`
		SELECT csrf_token, expires_at
		FROM sessions
		WHERE id = ?
	`, cookie.Value).Scan(&token, &expires)

	if err != nil || token == "" || time.Now().Unix() > expires {
		return "", false
	}

	return token, true
}

func requireCSRF(r *http.Request) bool {
	cookie, err := r.Cookie(sessionName)
	if err != nil {
		return false
	}

	token := r.FormValue("csrf_token")
	if token == "" {
		return false
	}

	var expected string

	err = db.QueryRow(`
		SELECT csrf_token
		FROM sessions
		WHERE id = ?
	`, cookie.Value).Scan(&expected)

	if err != nil || expected == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(expected),
		[]byte(token),
	) == 1
}

func validLibraryComponent(s string) bool {
	if s == "" {
		return false
	}

	if s == "." || s == ".." {
		return false
	}

	return !strings.ContainsAny(s, `/\`)
}

func handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if !verifyTurnstile(r) {
		http.Error(w, "CAPTCHA inválido", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var err error
	var userID int64
	var hash string

	err = db.QueryRow(
		`SELECT id, hash FROM users WHERE username = ?`, username,
	).Scan(&userID, &hash)

	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	/* Nesse ponto a sessão já foi validada, a senha tá certa */
	sessionID, _, err := newSession(userID)
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDevelopment,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((2 * time.Hour).Seconds()),
		Expires:  time.Now().Add(2 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"IP":               r.RemoteAddr,
		"UserAgent":        r.UserAgent(),
		"TurnstileSiteKey": os.Getenv("CLOUDFLARE_TURNSTILE_SITE_KEY"),
	}

	renderTemplate(w, "login.html", data)
}

func scanLibrary(root string) ([]Manga, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var mangas []Manga

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		mangaDir := filepath.Join(root, e.Name())

		vols, err := filepath.Glob(
			filepath.Join(mangaDir, "*.pdf"),
		)

		if err != nil || len(vols) == 0 {
			continue
		}

		sort.Strings(vols)

		names := make([]string, len(vols))

		for i, v := range vols {
			names[i] = filepath.Base(v)
		}

		cover := ""

		coverPath := filepath.Join(
			mangaDir,
			"cover.jpg",
		)

		if _, err := os.Stat(coverPath); err == nil {
			cover = "/cover?title=" + e.Name()
		}

		description := ""

		synopsisPath := filepath.Join(
			mangaDir,
			"sinopse.txt",
		)

		if data, err := os.ReadFile(synopsisPath); err == nil {
			description = strings.TrimSpace(
				string(data),
			)
		}

		mangas = append(
			mangas,
			Manga{
				Title:       e.Name(),
				Volumes:     names,
				Cover:       cover,
				Description: description,
			},
		)
	}

	sort.Slice(
		mangas,
		func(i, j int) bool {
			return strings.ToLower(
				mangas[i].Title,
			) < strings.ToLower(
				mangas[j].Title,
			)
		},
	)

	return mangas, nil
}

func handleCover(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	if title == "" {
		http.Error(
			w,
			"Parâmetros inválidos",
			http.StatusBadRequest,
		)
		return
	}

	if strings.Contains(title, "..") ||
		strings.ContainsAny(title, "/\\") {
		http.Error(
			w,
			"Acesso negado",
			http.StatusForbidden,
		)
		return
	}

	dataDirAbs, err := filepath.Abs(dataDir)
	if err != nil {
		http.Error(
			w,
			"Erro interno",
			http.StatusInternalServerError,
		)
		return
	}

	target := filepath.Join(
		dataDirAbs,
		title,
		"cover.jpg",
	)

	rel, err := filepath.Rel(
		dataDirAbs,
		target,
	)

	if err != nil ||
		rel == ".." ||
		strings.HasPrefix(
			rel,
			".."+string(os.PathSeparator),
		) {
		http.Error(
			w,
			"Acesso negado",
			http.StatusForbidden,
		)
		return
	}

	http.ServeFile(
		w,
		r,
		target,
	)
}
