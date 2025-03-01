package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// User 用户信息结构体
type User struct {
	UserID string
	APIKey string
}

// AccessLog 日志记录结构体
type AccessLog struct {
	UserID    string
	Timestamp time.Time
	Endpoint  string
}

// 数据库连接
var db *sql.DB

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		return err
	}

	// 创建用户表
	createUserTable := `
	CREATE TABLE IF NOT EXISTS proxy_user (
		user_id TEXT PRIMARY KEY,
		user_name TEXT NOT NULL,
		api_key TEXT NOT NULL
	);`
	_, err = db.Exec(createUserTable)
	if err != nil {
		return err
	}

	// 创建日志表
	createLogTable := `
	CREATE TABLE IF NOT EXISTS proxy_access_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		endpoint TEXT NOT NULL
	);`
	_, err = db.Exec(createLogTable)
	if err != nil {
		return err
	}

	// 创建模型表
	createProvidersTable := `
	CREATE TABLE IF NOT EXISTS proxy_provider (
		provider_name TEXT PRIMARY KEY,
		api_url TEXT NOT NULL,
		api_key TEXT NOT NULL
	);`
	_, err = db.Exec(createProvidersTable)
	if err != nil {
		return err
	}

	return nil
}

// 验证用户API Key
func validateUser(apiKey string) (string, bool) {
	var userID string
	err := db.QueryRow("SELECT user_id FROM proxy_user WHERE api_key = ?", apiKey).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false
		}
		log.Printf("Error querying user: %v", err)
		return "", false
	}
	return userID, true
}

// 记录访问日志
func logAccess(userID, endpoint string) {
	_, err := db.Exec("INSERT INTO proxy_access_log (user_id, endpoint) VALUES (?, ?)", userID, endpoint)
	if err != nil {
		log.Printf("Error logging access: %v", err)
	}
}

// 代理服务
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 获取请求中的API Key
	apiKey := r.Header.Get("Authorization")
	if apiKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 去除Bearer前缀
	apiKey = strings.TrimSpace(strings.TrimPrefix(apiKey, "Bearer "))

	// 验证API Key
	userID, valid := validateUser(apiKey)
	if !valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var path = r.URL.Path
	var provider = path[1:]
	var endIndex = strings.Index(provider, "/")
	if endIndex > 0 {
		provider = provider[0:endIndex]
	}
	log.Printf("Provider is %s", provider)

	// 查询模型信息
	var apiURL, modelAPIKey string
	err := db.QueryRow("SELECT api_url, api_key FROM proxy_provider WHERE provider_name = ?", provider).
		Scan(&apiURL, &modelAPIKey)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}
		log.Printf("Error querying provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 记录访问日志
	logAccess(userID, r.URL.Path)

	// 转发请求到目标提供者
	resp, err := forwardRequest(r, apiURL, modelAPIKey, strings.TrimPrefix(path, "/"+provider))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 将响应返回给客户端
	copyResponse(w, resp)
}

// 转发请求到OpenAI
func forwardRequest(r *http.Request, apiURL, apiKey string, path string) (*http.Response, error) {
	// 创建新的请求
	req, err := http.NewRequest(r.Method, apiURL+path, r.Body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range r.Header {
		req.Header[key] = value
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	return client.Do(req)
}

// 复制响应到客户端
func copyResponse(w http.ResponseWriter, resp *http.Response) {
	// 设置响应头
	for key, value := range resp.Header {
		w.Header().Set(key, strings.Join(value, ", "))
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 将响应体写入客户端
	io.Copy(w, resp.Body)
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	http.HandleFunc("/", proxyHandler)
	fmt.Println("Starting proxy server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
