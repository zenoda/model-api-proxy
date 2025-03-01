package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"time"
)

var db *sql.DB

func main() {
	// 初始化数据库
	initDB()

	// 定义命令行应用
	app := &cli.App{
		Name:  "proxy-admin",
		Usage: "Manage database for proxy users and model providers",
		Commands: []*cli.Command{
			{
				Name:    "add-user",
				Aliases: []string{"au"},
				Usage:   "Add a new user",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user-id", Required: true, Usage: "User's email as user id"},
					&cli.StringFlag{Name: "user-name", Required: true, Usage: "User Name"},
				},
				Action: addUser,
			},
			{
				Name:    "list-users",
				Aliases: []string{"lu"},
				Usage:   "List all users",
				Action:  listUsers,
			},
			{
				Name:    "delete-user",
				Aliases: []string{"du"},
				Usage:   "Delete a user by user ID",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "user-id", Required: true, Usage: "User ID"},
				},
				Action: deleteUser,
			},
			{
				Name:    "add-provider",
				Aliases: []string{"ap"},
				Usage:   "Add a new provider",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "provider-name", Required: true, Usage: "Provider Name"},
					&cli.StringFlag{Name: "api-url", Required: true, Usage: "API URL"},
					&cli.StringFlag{Name: "api-key", Required: true, Usage: "API Key"},
				},
				Action: addProvider,
			},
			{
				Name:    "list-providers",
				Aliases: []string{"lp"},
				Usage:   "List all providers",
				Action:  listProviders,
			},
			{
				Name:    "delete-provider",
				Aliases: []string{"dp"},
				Usage:   "Delete a provider by provider name",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "provider-name", Required: true, Usage: "Provider Name"},
				},
				Action: deleteProvider,
			}, {
				Name:  "list-logs",
				Usage: "View recent logs",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "lines",
						Value: 100,
						Usage: "Number of lines to view",
					},
					&cli.StringFlag{
						Name:  "user-id",
						Usage: "Filter logs by user ID",
					},
				},
				Action: viewLogs,
			},
			{
				Name:  "del-logs",
				Usage: "Delete old logs",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "days",
						Value: 30,
						Usage: "Number of days to retain logs",
					},
				},
				Action: deleteLogs,
			},
		},
	}

	// 运行应用
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initDB() {
	// 打开数据库（如果不存在则创建）
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
}

// 添加用户
func addUser(c *cli.Context) error {
	userID := c.String("user-id")
	userName := c.String("user-name")
	index := strings.Index(userID, "@")
	if index < 0 {
		return errors.New("please use user's email as user-id")
	}
	userNameEn := userID[0:index]

	apiKey := userNameEn + "-" + strings.ReplaceAll(uuid.New().String(), "-", "")

	_, err := db.Exec("INSERT INTO proxy_user (user_id, user_name, api_key) VALUES (?, ?, ?)", userID, userName, apiKey)
	if err != nil {
		log.Println("Error adding user:", err)
		return err
	}

	fmt.Println("User added successfully!")
	fmt.Println("--------------User Info-------------")
	fmt.Println("User ID:", userID)
	fmt.Println("User Name:", userName)
	fmt.Println("API Key:", apiKey)
	return nil
}

// 列出所有用户
func listUsers(c *cli.Context) error {
	rows, err := db.Query("SELECT user_id, user_name, api_key FROM proxy_user")
	if err != nil {
		log.Println("Error listing users:", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, userName, apiKey string
		if err := rows.Scan(&userID, &userName, &apiKey); err != nil {
			log.Println("Error scanning user:", err)
			return err
		}
		fmt.Printf("%s\t%s\t%s\n", userID, userName, apiKey)
	}

	return nil
}

// 删除用户
func deleteUser(c *cli.Context) error {
	userID := c.String("user-id")

	_, err := db.Exec("DELETE FROM proxy_user WHERE user_id = ?", userID)
	if err != nil {
		log.Println("Error deleting user:", err)
		return err
	}

	fmt.Println("User deleted successfully!")
	return nil
}

// 添加模型提供者
func addProvider(c *cli.Context) error {
	providerName := c.String("provider-name")
	apiURL := c.String("api-url")
	apiKey := c.String("api-key")

	_, err := db.Exec("INSERT INTO proxy_provider (provider_name, api_url, api_key) VALUES (?, ?, ?)", providerName, apiURL, apiKey)
	if err != nil {
		log.Println("Error adding provider:", err)
		return err
	}

	fmt.Println("Provider added successfully!")
	return nil
}

// 列出所有模型提供者
func listProviders(c *cli.Context) error {
	rows, err := db.Query("SELECT provider_name, api_url, api_key FROM proxy_provider")
	if err != nil {
		log.Println("Error listing providers:", err)
		return err
	}
	defer rows.Close()

	fmt.Println("Provider Name\tAPI URL\tAPI Key")
	for rows.Next() {
		var providerName, apiURL, apiKey string
		if err := rows.Scan(&providerName, &apiURL, &apiKey); err != nil {
			log.Println("Error scanning provider:", err)
			return err
		}
		fmt.Printf("%s\t%s\t%s\n", providerName, apiURL, apiKey)
	}

	return nil
}

// 删除模型提供者
func deleteProvider(c *cli.Context) error {
	providerName := c.String("provider-name")

	_, err := db.Exec("DELETE FROM proxy_provider WHERE provider_name = ?", providerName)
	if err != nil {
		log.Println("Error deleting provider:", err)
		return err
	}

	fmt.Println("Provider deleted successfully!")
	return nil
}

// 查看日志
func viewLogs(c *cli.Context) error {
	lines := c.Int("lines")
	user := c.String("user-id")

	query := "SELECT user_id, timestamp, endpoint FROM proxy_access_log ORDER BY timestamp DESC LIMIT ?"
	args := []interface{}{lines}

	if user != "" {
		query = "SELECT user_id, timestamp, endpoint FROM proxy_access_log WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?"
		args = []interface{}{user, lines}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Println("Error querying logs:", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var timestamp time.Time
		var endpoint string
		if err := rows.Scan(&userID, &timestamp, &endpoint); err != nil {
			log.Println("Error scanning log:", err)
			return err
		}
		fmt.Printf("%s\t%s\t%s\n", userID, timestamp.In(time.Local).Format("2006-01-02 15:04:05"), endpoint)
	}

	return nil
}

// 删除日志
func deleteLogs(c *cli.Context) error {
	days := c.Int("days")

	if days <= 0 {
		log.Println("Error: days must be a positive integer")
		return fmt.Errorf("days must be a positive integer")
	}

	// 计算保留日志的截止时间
	retainUntil := time.Now().AddDate(0, 0, -days)

	_, err := db.Exec("DELETE FROM proxy_access_log WHERE timestamp < ?", retainUntil)
	if err != nil {
		log.Println("Error deleting logs:", err)
		return err
	}

	fmt.Printf("Deleted logs older than %d days\n", days)
	return nil
}
