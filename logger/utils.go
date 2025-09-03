package logger

import (
	"github.com/google/uuid"
)

// GenerateRequestID 生成唯一的请求ID
// 返回: string 生成的UUID字符串
func GenerateRequestID() string {
	return uuid.New().String()
}

// NewZapFactory 创建zap日志工厂实例
// 返回: Factory zap日志工厂
func NewZapFactory() Factory {
	return &ZapFactory{}
}