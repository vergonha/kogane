package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	addr          string
	r2Client      *s3.Client
	r2Bucket      string
	cachedLibrary []Manga
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

	_, err = db.Exec(`
	    INSERT INTO users (username, hash)
	    VALUES ('admin', ?)
	    ON CONFLICT(username) DO UPDATE SET hash = excluded.hash
	`, string(hash))

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

	r2Bucket = os.Getenv("R2_BUCKET_NAME")
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")

	if r2Bucket == "" || accountID == "" || accessKey == "" || secretKey == "" {
		log.Fatal("R2_BUCKET_NAME, R2_ACCOUNT_ID, R2_ACCESS_KEY_ID e R2_SECRET_ACCESS_KEY devem estar configuradas")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Erro ao carregar config R2: %v", err)
	}

	// Instancia o cliente apontando diretamente para o endpoint do R2
	r2URL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	r2Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2URL)
	})

	libData, err := os.ReadFile("library.json")
	if err != nil {
		log.Fatalf("Erro ao ler library.json: %v", err)
	}
	if err := json.Unmarshal(libData, &cachedLibrary); err != nil {
		log.Fatalf("Erro ao parsear library.json: %v", err)
	}

	db, err = sql.Open("sqlite3", "file:manga.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		log.Fatal(err)
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
	log.Fatal(http.ListenAndServe(addr, mux))
}

func handlePDF(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	vol := r.URL.Query().Get("vol")

	if !validLibraryComponent(title) || !validLibraryComponent(vol) {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(strings.ToLower(vol), ".pdf") {
		http.Error(w, "Parâmetros inválidos", http.StatusBadRequest)
		return
	}

	key := title + "/" + vol

	presignClient := s3.NewPresignClient(r2Client)

	presignReq, err := presignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(r2Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		http.Error(w, "Erro ao gerar link de download", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignReq.URL, http.StatusTemporaryRedirect)
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
	csrfToken, ok := getCSRFToken(r)
	if !ok {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	renderTemplate(w, "dashboard.html", map[string]any{
		"Mangas":    cachedLibrary,
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

	return !strings.Contains(s, "..") && !strings.ContainsAny(s, `/\`)
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
	ip := r.Header.Get("CF-Connecting-IP")

	data := map[string]string{
		"IP":               ip,
		"UserAgent":        r.UserAgent(),
		"TurnstileSiteKey": os.Getenv("CLOUDFLARE_TURNSTILE_SITE_KEY"),
	}

	renderTemplate(w, "login.html", data)
}

func handleCover(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")

	if !validLibraryComponent(title) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	key := title + "/cover.jpg"

	presignClient := s3.NewPresignClient(r2Client)

	presignReq, err := presignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(r2Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(1*time.Hour))

	if err != nil {
		log.Printf("erro ao gerar link da imagem: %v", err)
		http.Error(w, "Erro ao gerar link da imagem", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignReq.URL, http.StatusTemporaryRedirect)
}
