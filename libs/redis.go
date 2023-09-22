package libs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gmf001/go-microservice/models"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func orderIDKey(orderID uint64) string {
	return fmt.Sprintf("order:%d", orderID)
}

func (r *RedisClient) Insert(ctx context.Context, order models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDKey(order.OrderID)

	txn := r.Client.TxPipeline()

	res := r.Client.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if err := r.Client.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}

var ErrNotExists = errors.New("order does not exist")

func (r *RedisClient) FindByID(ctx context.Context, id uint64) (models.Order, error) {
	key := orderIDKey(id)

	data, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return models.Order{}, ErrNotExists
	} else if err != nil {
		return models.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	var order models.Order
	err = json.Unmarshal([]byte(data), &order)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to decode order: %w", err)
	}

	return order, nil
}

func (r *RedisClient) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExists
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("failed to get order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (r *RedisClient) Update(ctx context.Context, order models.Order) error  {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to encode order: %w", err)
	}

	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExists
	} else if err != nil {
		return fmt.Errorf("UPDATE: failed to update order: %w", err)
	}

	return nil
}

type FindAllOptions struct {
	Limit  uint64
	Offset uint64
}

type FindResult struct {
	Orders []models.Order
	Cursor  uint64
}

func (r *RedisClient) FindAll(ctx context.Context, options FindAllOptions) (FindResult, error)  {
	res := r.Client.SScan(ctx, "orders", options.Offset, "", int64(options.Limit))
	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []models.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]models.Order, 0, len(xs))

	for _, x := range xs {
		var order models.Order
		err := json.Unmarshal([]byte(x.(string)), &order)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode order: %w", err)
		}

		orders = append(orders, order)
	}

	return FindResult{
		Orders: orders,
		Cursor: uint64(cursor),
	}, nil
}