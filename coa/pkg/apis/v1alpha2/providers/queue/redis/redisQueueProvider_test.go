package redisqueue

import (
	"context"
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRedisClient() *redis.Client {
	options := &redis.Options{
		Addr: "localhost:6380",
	}
	client := redis.NewClient(options)
	return client
}

func TestRedisQueueProvider_Enqueue(t *testing.T) {
	client := setupRedisClient()
	queueName := "test_queue"
	rq := NewRedisQueue(client, queueName)
	rq.Ctx = context.TODO()

	element := map[string]interface{}{"key": "value"}
	messageID, err := rq.Enqueue(queueName, element, context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, messageID)
}

func TestRedisQueueProvider_Peek(t *testing.T) {
	client := setupRedisClient()
	queueName := "test_queue"
	rq := NewRedisQueue(client, queueName)
	rq.Ctx = context.TODO()

	element := map[string]interface{}{"key": "value"}
	_, err := rq.Enqueue(queueName, element, context.Background())
	assert.NoError(t, err)

	result, err := rq.Peek(queueName, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, element, result)
}

func TestRedisQueueProvider_Dequeue(t *testing.T) {
	client := setupRedisClient()
	queueName := "test_queue"
	rq := NewRedisQueue(client, queueName)
	rq.Ctx = context.TODO()

	element := map[string]interface{}{"key": "value"}
	_, err := rq.Enqueue(queueName, element, context.Background())
	assert.NoError(t, err)

	result, err := rq.Dequeue(queueName, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, element, result)
}

func TestRedisQueueProvider_Size(t *testing.T) {
	client := setupRedisClient()
	queueName := "test_queue"
	rq := NewRedisQueue(client, queueName)
	rq.Ctx = context.TODO()

	for i := 0; i < 10; i++ {
		element := map[string]interface{}{"key": "value_" + strconv.Itoa(i)}
		_, err := rq.Enqueue(queueName, element, context.Background())
		assert.NoError(t, err)
	}

	size := rq.Size(queueName, context.Background())
	assert.Equal(t, 10, size)
}

func TestRedisQueueProvider_QueryByPaging(t *testing.T) {
	client := setupRedisClient()
	queueName := "test_queue"
	rq := NewRedisQueue(client, queueName)
	rq.Ctx = context.TODO()

	// 插入测试数据
	for i := 0; i < 10; i++ {
		element := map[string]interface{}{"key": "value_" + strconv.Itoa(i)}
		_, err := rq.Enqueue(queueName, element, context.Background())
		assert.NoError(t, err)
	}

	// 测试 QueryByPaging 方法
	results, lastMessageID, err := rq.QueryByPaging(queueName, "0", 5, context.Background())
	assert.NoError(t, err)

	// 预期结果
	expectedResults := make([]map[string]interface{}, 5)
	for i := 0; i < 5; i++ {
		expectedResults[i] = map[string]interface{}{"key": "value_" + strconv.Itoa(i)}
	}

	results, lastMessageID, err = rq.QueryByPaging(queueName, lastMessageID, 2, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, len(results))

	// 验证结果
	assert.Equal(t, expectedResults, results)
	assert.NotEmpty(t, lastMessageID)
}
