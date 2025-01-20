package redisqueue

import (
	"context"
	"redis"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestQueryByPaging(t *testing.T) {
	db, mock := redismock.NewClientMock()
	rq := &RedisQueueProvider{
		client: db,
		Ctx:    context.TODO(),
	}

	queueName := "testQueue"
	start := "0"
	size := 2

	mock.ExpectXRangeN(queueName, "0", "+", int64(size+1)).SetVal([]redis.XMessage{
		{ID: "1", Values: map[string]interface{}{"data": "message1"}},
		{ID: "2", Values: map[string]interface{}{"data": "message2"}},
		{ID: "3", Values: map[string]interface{}{"data": "message3"}},
	})

	results, lastID, err := rq.QueryByPaging(queueName, start, size)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "2", lastID)
	assert.Equal(t, []byte("message1"), results[0])
	assert.Equal(t, []byte("message2"), results[1])

	err = mock.ExpectationsWereMet()
	assert.Nil(t, err)
}
