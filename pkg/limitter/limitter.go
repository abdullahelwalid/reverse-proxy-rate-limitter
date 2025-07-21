package limitter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func GetClient() {
	fmt.Println("Initilaizing Redis Connection")
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})
	ctx := context.TODO()
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Connected to Redis")
	RedisClient = client
}

type ClientToken struct {
	IPAddr string
	Tokens int
}

type ErrorClientTokenLimitExceed struct {
	Message string
}

func (e *ErrorClientTokenLimitExceed) Error() string {
	return e.Message
}

type ErrorRedis struct {
	Message string
}

func (e *ErrorRedis) Error() string {
	return e.Message
}

func (c *ClientToken) Init(IPAddr string) *ClientToken {
	c.IPAddr = IPAddr
	c.Tokens = 10 // Default tokens
	return c
}

func (c *ClientToken) Save() error {
	err := RedisClient.Set(context.TODO(), c.IPAddr, c.Tokens, time.Minute*1).Err()
	return err
}

// Every API call will call Consume() method to reduce Tokens -=1
func (c *ClientToken) Consume() error {
	token, err := RedisClient.Get(context.TODO(), c.IPAddr).Result()
	if errors.Is(err, redis.Nil) {
		fmt.Println("No record found")
		c.Tokens = 20
		c.Save()
		return nil
	} else if err != nil {
		return &ErrorRedis{Message: "Error Redis"}
	}
	tokenInt, err := strconv.Atoi(token)
	if err != nil {
		return err
	}
	// Check if all tokens are consumed
	if tokenInt == 0 {
		return &ErrorClientTokenLimitExceed{Message: "Limit Exceeded"}
	}
	// Consume token with current TTL
	RedisClient.Set(context.TODO(), c.IPAddr, tokenInt-1, redis.KeepTTL)
	return err
}

// Retrieve Client Token from Redis
func (c *ClientToken) Get(IPAddr string) *ClientToken {
	return c
}
