package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/swayanshu-2003/classroom-backend/middlewares"
	"github.com/swayanshu-2003/classroom-backend/models"
	"github.com/swayanshu-2003/classroom-backend/storage"
)

func main() {
	fmt.Println("welcome to classroom backend")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Password: os.Getenv("DB_PASSWORD"),
		Username: os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_DATABASE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateUser(db)

	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := middlewares.Repository{
		DB: db,
	}

	app := fiber.New()

	r.SetupRoutes(app)

	log.Fatal(app.Listen(":5600"))

}
