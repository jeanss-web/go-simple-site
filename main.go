package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

// Загрузка всех шаблонов
var (
	templates = template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/catalog.html",
		"templates/sale.html",
		"templates/contact.html",
	))
	db *sql.DB
)

type Product struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Обработчик для каждой страницы
func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT name, price FROM products")
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	products := []Product{}

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Name, &p.Price); err != nil {
			http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
			return
		}
		products = append(products, p)
	}

	templates.ExecuteTemplate(w, "catalog.html", products)
}

func saleHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "sale.html", nil)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "contact.html", nil)
}

func createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price NUMERIC(10,2) NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}

func addProduct(name string, price float64) error {
	_, err := db.Exec("INSERT INTO products (name, price) VALUES ($1, $2)", name, price)
	return err
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	err := addProduct(p.Name, p.Price)
	if err != nil {
		http.Error(w, "Ошибка базы данных", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Продукт успешно добавлен")
}

func main() {

	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:1111@localhost:5432/fashionstore?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal("Ошибка при открытии базы данных:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Ошибка при подключении к базе данных:", err)
	}

	if err = createTable(); err != nil {
		log.Fatal("Ошибка при создании таблицы:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("Сервер запущен на порту " + port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/catalog", catalogHandler)
	mux.HandleFunc("/sale", saleHandler)
	mux.HandleFunc("/contact", contactHandler)
	mux.HandleFunc("/add-product", productHandler)
	http.ListenAndServe(":"+port, mux)
}
