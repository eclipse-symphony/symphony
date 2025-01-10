package redisqueue

import (
	"os"
	"testing"

	"github.com/eclipse-symphony/symphony/coa/pkg/apis/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestWithEmptyConfig(t *testing.T) {
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestWithMissingHost(t *testing.T) {
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{
		Name:     "test",
		Password: "abc",
	})
	assert.NotNil(t, err)
	coaErr, ok := err.(v1alpha2.COAError)
	assert.True(t, ok)
	assert.Equal(t, v1alpha2.MissingConfig, coaErr.State)
}

func TestInit(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS environment variable is not set")
	}
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{
		Name:     "test",
		Host:     "abc",
		Password: "",
	})
	assert.Nil(t, err)
}

func TestInitWithMap(t *testing.T) {
	provider := RedisQueueProvider{}
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis != "" {
		err := provider.InitWithMap(
			map[string]string{
				"name": "test",
				"host": "abc",
			},
		)
		assert.Nil(t, err) // Provider initialization succeeds if redis is running
	}

	err := provider.InitWithMap(
		map[string]string{
			"name": "test",
		},
	)
	assert.NotNil(t, err)
	err = provider.InitWithMap(
		map[string]string{
			"name":        "test",
			"host":        "abc",
			"requiresTLS": "abcd",
		},
	)
	assert.NotNil(t, err)
	err = provider.InitWithMap(
		map[string]string{
			"name": "test",
			"host": "abc",
		},
	)
	assert.NotNil(t, err)
}

func TestEnqueue(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS environment variable is not set")
	}
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{
		Name:     "test",
		Host:     "abc",
		Password: "",
	})
	assert.Nil(t, err)

	element := v1alpha2.Event{Body: "test"}
	id, err := provider.Enqueue("testQueue", element)
	assert.Nil(t, err)
	assert.NotEmpty(t, id)
}

func TestPeek(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS environment variable is not set")
	}
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{
		Name:     "test",
		Host:     "abc",
		Password: "",
	})
	assert.Nil(t, err)

	element := v1alpha2.Event{Body: "test"}
	_, err = provider.Enqueue("testQueue", element)
	assert.Nil(t, err)

	msg, err := provider.Peek("testQueue")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
}

func TestDequeue(t *testing.T) {
	testRedis := os.Getenv("TEST_REDIS")
	if testRedis == "" {
		t.Skip("Skipping because TEST_REDIS environment variable is not set")
	}
	provider := RedisQueueProvider{}
	err := provider.Init(RedisQueueProviderConfig{
		Name:     "test",
		Host:     "abc",
		Password: "",
	})
	assert.Nil(t, err)

	element := v1alpha2.Event{Body: "test"}
	_, err = provider.Enqueue("testQueue", element)
	assert.Nil(t, err)

	msg, err := provider.Dequeue("testQueue")
	assert.Nil(t, err)
	assert.NotNil(t, msg)
}

func TestRedisQueueProviderConfigFromMap(t *testing.T) {
	configMap := map[string]string{
		"name":        "test",
		"host":        "abc",
		"password":    "123",
		"requiresTLS": "true",
	}
	config, err := RedisQueueProviderConfigFromMap(configMap)
	assert.Nil(t, err)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, "abc", config.Host)
	assert.Equal(t, "123", config.Password)
	assert.Equal(t, true, config.RequiresTLS)
}
