package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/theFalconator/expenses/activity"
	"github.com/theFalconator/expenses/config"
	"github.com/theFalconator/expenses/groups"
	"github.com/theFalconator/expenses/pages"
	"github.com/theFalconator/expenses/users"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Incoming request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Println("app started")
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg[config.DATABASE_URL])
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	groupStore := groups.NewStore(db)
	userStore := users.NewStore(db)
	activityStore := activity.NewStore(db)
	indexPageHandler := pages.NewIndexPageHandler(userStore, activityStore, groupStore)
	addPageHandler := pages.NewAddExpenseHandler(userStore, activityStore)
	settleHandler := pages.NewSettlementHandler(userStore, activityStore)
	importHandler := pages.NewImportHandler(userStore)
	uploadHandler := pages.NewUploadHandler(activityStore, userStore)

	fs := http.FileServer(http.Dir("./static"))

	router := http.NewServeMux()
	router.Handle("/static/", http.StripPrefix("/static/", fs))
	router.Handle("/settle/{id}", Logging(settleHandler))
	router.Handle("/", Logging(indexPageHandler))
	router.Handle("/{id}/", Logging(indexPageHandler))
	router.Handle("/add-expense/{id}", Logging(addPageHandler))
	router.Handle("/import/{id}", Logging(importHandler))
	router.Handle("/upload", Logging(uploadHandler))

	server := http.Server{
		Addr:    ":7074",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
