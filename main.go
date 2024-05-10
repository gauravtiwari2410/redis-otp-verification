package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

var redisClient *redis.Client

func init() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator once
}

func main() {
	app := fiber.New()
	initRedis()
	app.Post("/generate_otp", generateOTP)
	app.Post("/verify_otp", verifyOTP)
	app.Listen(":3000")
}

func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Default DB
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

func generateOTP(c *fiber.Ctx) error {
	user := c.Query("user")
	if user == "" {
		return c.Status(400).SendString("User ID is required")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000)) // Generate a six-digit OTP
	err := redisClient.Set(context.Background(), user, otp, 5*time.Minute).Err()
	if err != nil {
		return c.Status(500).SendString("Failed to store OTP")
	}
	return c.SendString("OTP generated and sent")
}

func verifyOTP(c *fiber.Ctx) error {
	user := c.Query("user")
	otp := c.Query("otp")
	if user == "" || otp == "" {
		return c.Status(400).SendString("User and OTP are required")
	}

	storedOtp, err := redisClient.Get(context.Background(), user).Result()
	if err == redis.Nil {
		return c.Status(404).SendString("OTP expired or does not exist")
	} else if err != nil {
		return c.Status(500).SendString("Failed to retrieve OTP")
	}

	if storedOtp != otp {
		return c.Status(403).SendString("Invalid OTP")
	}

	return c.SendString("OTP verified successfully")
}
