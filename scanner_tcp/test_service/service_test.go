package test_service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestScannerServiceIntegration(t *testing.T) {
	// Конфигурация из переменных окружения
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	queueName := getEnv("SCANNER_NAME", "scanner1")
	testTimeout := 30 * time.Second

	t.Log("=== Начало интеграционного теста ===")
	t.Logf("Используем очередь: %s", queueName)
	t.Logf("RabbitMQ URL: %s", rabbitMQURL)

	// 1. Подключение к RabbitMQ с ретраями
	conn, err := connectRabbitMQWithRetry(rabbitMQURL, 3, 2*time.Second)
	if err != nil {
		t.Fatalf("Не удалось подключиться к RabbitMQ: %v", err)
	}
	defer conn.Close()
	t.Log("✓ Подключение к RabbitMQ успешно")

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Не удалось создать канал: %v", err)
	}
	defer ch.Close()

	// 2. Проверка существования целевой очереди
	if err := checkQueueExists(ch, queueName); err != nil {
		t.Fatalf("Ошибка проверки очереди: %v", err)
	}
	t.Logf("✓ Очередь %s существует", queueName)

	// 3. Создание временной очереди для ответов
	replyQueue, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("Не удалось создать очередь ответов: %v", err)
	}
	t.Logf("✓ Временная очередь ответов: %s", replyQueue.Name)

	// 4. Подписка на ответы
	msgs, err := ch.Consume(
		replyQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("Не удалось подписаться на очередь: %v", err)
	}
	t.Log("✓ Подписка на ответы активирована")

	// 5. Подготовка тестового запроса
	testReq := map[string]interface{}{
		"ip":     "google.com",
		"ports":  "80-60000",
		"taskId": fmt.Sprintf("test-%d", time.Now().Unix()),
	}
	reqBody, _ := json.Marshal(testReq)

	t.Log("Тестовый запрос:")
	t.Log(string(reqBody))

	// 6. Отправка запроса
	correlationID := fmt.Sprintf("corr-%s", testReq["taskId"])
	err = ch.PublishWithContext(
		context.Background(),
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			ReplyTo:       replyQueue.Name,
			Body:          reqBody,
			Timestamp:     time.Now(),
		},
	)
	if err != nil {
		t.Fatalf("Ошибка отправки запроса: %v", err)
	}
	t.Logf("✓ Запрос отправлен (CorrelationID: %s)", correlationID)

	// 7. Ожидание ответа
	t.Logf("Ожидание ответа (таймаут: %v)...", testTimeout)
	select {
	case msg := <-msgs:
		t.Log("=== Получен ответ ===")
		t.Logf("CorrelationID: %s", msg.CorrelationId)
		t.Log("Тело ответа:")
		t.Log(string(msg.Body))

		var response struct {
			TaskID    string `json:"taskId"`
			Status    string `json:"status"`
			OpenPorts []int  `json:"openPorts"`
			Error     string `json:"error,omitempty"`
		}

		if err := json.Unmarshal(msg.Body, &response); err != nil {
			t.Errorf("Ошибка декодирования ответа: %v", err)
		} else {
			t.Log("Декодированный ответ:")
			t.Logf("- TaskID: %s", response.TaskID)
			t.Logf("- Status: %s", response.Status)
			t.Logf("- OpenPorts: %v", response.OpenPorts)
			if response.Error != "" {
				t.Logf("- Error: %s", response.Error)
			}
		}

		if msg.CorrelationId != correlationID {
			t.Errorf("Несоответствие CorrelationID: ожидалось %s, получено %s",
				correlationID, msg.CorrelationId)
		}

	case <-time.After(testTimeout):
		t.Fatal("Таймаут ожидания ответа")
	}

	t.Log("=== Тест завершен успешно ===")
}

// Вспомогательные функции

func connectRabbitMQWithRetry(url string, maxRetries int, interval time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("после %d попыток: %v", maxRetries, err)
}

func checkQueueExists(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclarePassive(
		queueName,
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	return err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
