package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"flightapi/internal/repository/postgres"
	"flightapi/internal/swaggerui"
	"flightapi/internal/transport/httpapi"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@127.0.0.1:5432/demo?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := pgxpool.New(ctx, dsn)
	cancel()
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	pctx, pcancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = pool.Ping(pctx)
	pcancel()
	if err != nil {
		log.Fatalf("postgres ping: %v", err)
	}
	log.Println("connected to PostgreSQL")

	repo := postgres.New(pool)
	handler := httpapi.NewHandler(repo)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware())
	r.GET("/docs", gin.WrapF(swaggerui.DocsHandler))
	r.GET("/openapi.yaml", gin.WrapF(swaggerui.OpenAPIYAMLHandler))
	handler.RegisterRoutes(r)

	addr := ":8080"
	log.Printf("Flight API server listening on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Accept-Language")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
