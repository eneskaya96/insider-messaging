package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/eneskaya/insider-messaging/internal/application/dto"
	"github.com/eneskaya/insider-messaging/internal/application/service"
	"github.com/eneskaya/insider-messaging/internal/domain/entity"
	"github.com/eneskaya/insider-messaging/internal/domain/repository"
	"github.com/eneskaya/insider-messaging/internal/domain/valueobject"
	"github.com/eneskaya/insider-messaging/internal/infrastructure/cache"
	infrahttp "github.com/eneskaya/insider-messaging/internal/infrastructure/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, msg *entity.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageRepository) Update(ctx context.Context, msg *entity.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) FindSentMessages(ctx context.Context, limit, offset int) ([]*entity.Message, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) FindPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) GetStats(ctx context.Context) (*repository.MessageStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.MessageStats), args.Error(1)
}

func (m *MockMessageRepository) BeginTx(ctx context.Context) (repository.Transaction, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Transaction), args.Error(1)
}

// Mock Transaction
type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransaction) GetContext() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

// Mock Webhook Client
type MockWebhookClient struct {
	mock.Mock
}

func (m *MockWebhookClient) SendMessage(ctx context.Context, phone, content string) (*infrahttp.WebhookResponse, error) {
	args := m.Called(ctx, phone, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*infrahttp.WebhookResponse), args.Error(1)
}

// Mock Cache
type MockMessageCache struct {
	mock.Mock
}

func (m *MockMessageCache) CacheSentMessage(ctx context.Context, msg *cache.CachedMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageCache) GetSentMessage(ctx context.Context, messageID string) (*cache.CachedMessage, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cache.CachedMessage), args.Error(1)
}

func (m *MockMessageCache) IsCached(ctx context.Context, messageID string) (bool, error) {
	args := m.Called(ctx, messageID)
	return args.Bool(0), args.Error(1)
}

// Tests
func TestCreateMessage_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	req := &dto.CreateMessageRequest{
		PhoneNumber: "+905551234567",
		Content:     "Test message",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Message")).
		Return(nil)

	// Act
	result, err := svc.CreateMessage(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.PhoneNumber, result.PhoneNumber)
	assert.Equal(t, req.Content, result.Content)
	assert.Equal(t, "pending", result.Status)
	assert.Equal(t, 0, result.Attempts)
	assert.Equal(t, 3, result.MaxAttempts)
	mockRepo.AssertExpectations(t)
}

func TestCreateMessage_InvalidPhone(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	req := &dto.CreateMessageRequest{
		PhoneNumber: "invalid-phone",
		Content:     "Test",
	}

	// Act
	result, err := svc.CreateMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "phone number")
}

func TestCreateMessage_EmptyContent(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	req := &dto.CreateMessageRequest{
		PhoneNumber: "+905551234567",
		Content:     "",
	}

	// Act
	result, err := svc.CreateMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "content")
}

func TestCreateMessage_ContentTooLong(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	// Create a string with 161 'a' characters
	longContent := ""
	for i := 0; i < 161; i++ {
		longContent += "a"
	}

	req := &dto.CreateMessageRequest{
		PhoneNumber: "+905551234567",
		Content:     longContent, // 161 characters (exceeds 160 limit)
	}

	// Act
	result, err := svc.CreateMessage(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetMessage_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	messageID := uuid.New()
	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test", 160)
	message, _ := entity.NewMessage(phone, content, 3)

	mockRepo.On("FindByID", mock.Anything, messageID).Return(message, nil)

	// Act
	result, err := svc.GetMessage(context.Background(), messageID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "+905551234567", result.PhoneNumber)
	mockRepo.AssertExpectations(t)
}

func TestGetMessage_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	messageID := uuid.New()
	mockRepo.On("FindByID", mock.Anything, messageID).Return(nil, errors.New("not found"))

	// Act
	result, err := svc.GetMessage(context.Background(), messageID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestProcessPendingMessages_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test message", 160)
	message, _ := entity.NewMessage(phone, content, 3)

	mockTx := new(MockTransaction)
	mockRepo.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockTx.On("GetContext").Return(context.Background())
	mockRepo.On("FindPendingMessages", mock.Anything, 10).
		Return([]*entity.Message{message}, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Message")).
		Return(nil)

	webhookResp := &infrahttp.WebhookResponse{
		MessageID: "webhook-123",
		Message:   "Message sent successfully",
	}
	mockWebhook.On("SendMessage", mock.Anything, "+905551234567", "Test message").
		Return(webhookResp, nil)

	mockCache.On("CacheSentMessage", mock.Anything, mock.AnythingOfType("*cache.CachedMessage")).
		Return(nil)
	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	// Act
	count, err := svc.ProcessPendingMessages(context.Background(), 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
	mockCache.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestProcessPendingMessages_NoMessages(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	mockTx := new(MockTransaction)
	mockRepo.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockTx.On("GetContext").Return(context.Background())
	mockRepo.On("FindPendingMessages", mock.Anything, 10).
		Return([]*entity.Message{}, nil)
	mockTx.On("Rollback").Return(nil)

	// Act
	count, err := svc.ProcessPendingMessages(context.Background(), 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestProcessPendingMessages_WebhookFailure(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test", 160)
	message, _ := entity.NewMessage(phone, content, 3)

	mockTx := new(MockTransaction)
	mockRepo.On("BeginTx", mock.Anything).Return(mockTx, nil)
	mockTx.On("GetContext").Return(context.Background())
	mockRepo.On("FindPendingMessages", mock.Anything, 10).
		Return([]*entity.Message{message}, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Message")).
		Return(nil).Times(2) // Once for processing, once for failed

	mockWebhook.On("SendMessage", mock.Anything, "+905551234567", "Test").
		Return(nil, errors.New("webhook error"))

	mockTx.On("Commit").Return(nil)
	mockTx.On("Rollback").Return(nil)

	// Act
	count, err := svc.ProcessPendingMessages(context.Background(), 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count) // Failed messages don't count
	mockRepo.AssertExpectations(t)
	mockWebhook.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestGetSentMessages_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	phone, _ := valueobject.NewPhoneNumber("+905551234567")
	content, _ := valueobject.NewMessageContent("Test", 160)
	message1, _ := entity.NewMessage(phone, content, 3)
	message2, _ := entity.NewMessage(phone, content, 3)

	stats := &repository.MessageStats{
		TotalMessages:   10,
		SentMessages:    2,
		FailedMessages:  3,
		PendingMessages: 5,
	}

	mockRepo.On("FindSentMessages", mock.Anything, 20, 0).
		Return([]*entity.Message{message1, message2}, nil)
	mockRepo.On("GetStats", mock.Anything).Return(stats, nil)

	// Act (page=1, pageSize=20)
	result, err := svc.GetSentMessages(context.Background(), 1, 20)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Messages, 2)
	assert.Equal(t, "+905551234567", result.Messages[0].PhoneNumber)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	mockRepo.AssertExpectations(t)
}

func TestGetSentMessages_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	stats := &repository.MessageStats{
		TotalMessages:   0,
		SentMessages:    0,
		FailedMessages:  0,
		PendingMessages: 0,
	}

	mockRepo.On("FindSentMessages", mock.Anything, 20, 0).
		Return([]*entity.Message{}, nil)
	mockRepo.On("GetStats", mock.Anything).Return(stats, nil)

	// Act
	result, err := svc.GetSentMessages(context.Background(), 1, 20)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Messages)
	assert.Equal(t, 0, result.TotalCount)
	mockRepo.AssertExpectations(t)
}

func TestGetStats_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	stats := &repository.MessageStats{
		TotalMessages:   100,
		SentMessages:    80,
		FailedMessages:  15,
		PendingMessages: 5,
	}

	mockRepo.On("GetStats", mock.Anything).Return(stats, nil)

	// Act
	result, err := svc.GetStats(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.TotalMessages)
	assert.Equal(t, int64(80), result.SentMessages)
	assert.Equal(t, int64(15), result.FailedMessages)
	assert.Equal(t, int64(5), result.PendingMessages)
	mockRepo.AssertExpectations(t)
}

func TestGetStats_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMessageRepository)
	mockWebhook := new(MockWebhookClient)
	mockCache := new(MockMessageCache)

	svc := service.NewMessageService(mockRepo, mockWebhook, mockCache, 160, 3)

	mockRepo.On("GetStats", mock.Anything).Return(nil, errors.New("database error"))

	// Act
	result, err := svc.GetStats(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}
