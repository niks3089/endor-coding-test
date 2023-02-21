package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type ObjectDB interface {
	// Store will store the object in the data store. The object will have a
	// name and kind, and the Store method should create a unique ID.
	Store(ctx context.Context, object Object) error
	// GetObjectByID will retrieve the object with the provided ID.
	GetObjectByID(ctx context.Context, id string) (Object, error)
	// GetObjectsByName will retrieve the object with the given name.
	GetObjectsByName(ctx context.Context, name string) ([]Object, error)
	// ListObjects will return a list of all objects of the given kind.
	ListObjects(ctx context.Context, kind string) ([]Object, error)
	// DeleteObject will delete the object.
	DeleteObject(ctx context.Context, id string) error
}

type RedisDB struct {
	client *redis.Client
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func (db *RedisDB) listObjects(pattern string) ([]Object, error) {
	var objects []Object
	var object Object

	keys, err := db.client.Keys(pattern).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return objects, nil
	}

	for _, key := range keys {
		value, err := db.client.Get(key).Result()
		if err != nil {
			return nil, err
		}
		object, err = extractObject(keys[0], value)
		if err != nil {
			return nil, err
		}

		objects = append(objects, object)
	}
	return objects, nil
}

func NewRedisDB(addr, password string) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisDB{client}, nil
}

func (db *RedisDB) Store(ctx context.Context, object Object) error {
	id := uuid.New().String()

	value, err := json.Marshal(object)
	if err != nil {
		return err
	}

	object.SetID(id)
	key := id + "::" + object.GetName() + "::" + object.GetKind()

	if strings.Count(key, "::") != 2 {
		return errors.New("unexpected restricted delimiter in the key")
	}

	err = db.client.Set(key, value, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (db *RedisDB) GetObjectByID(ctx context.Context, id string) (Object, error) {
	pattern := fmt.Sprintf("%s::*::*", id)
	keys, err := db.client.Keys(pattern).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, errors.New("object not found")
	}

	if len(keys) > 1 {
		return nil, errors.New("multiple objects with the same ID")
	}

	value, err := db.client.Get(keys[0]).Result()
	if err != nil {
		return nil, err
	}

	return extractObject(keys[0], value)
}

func (db *RedisDB) GetObjectsByName(ctx context.Context, name string) ([]Object, error) {
	if name == "" {
		return nil, errors.New("invalid request. empty name")
	}

	return db.listObjects(fmt.Sprintf("*::%s::*", name))
}

func (db *RedisDB) ListObjects(ctx context.Context, kind string) ([]Object, error) {
	if kind == "" {
		return nil, errors.New("invalid request. empty kind")
	}

	return db.listObjects(fmt.Sprintf("*::*::%s", kind))
}

func (db *RedisDB) DeleteObject(ctx context.Context, id string) error {
	pattern := fmt.Sprintf("%s::*::*", id)
	keys, err := db.client.Keys(pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}
	if len(keys) > 1 {
		return errors.New("multiple objects with the same ID")
	}

	key := keys[0]

	_, err = db.client.Del(key).Result()
	if err != nil {
		return err
	}
	return nil
}
