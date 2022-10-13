package system

import (
	"context"

	"lovebox/models"
	"lovebox/pkg/database"

	"go.uber.org/zap"
)

type Service struct {
	log         *zap.SugaredLogger
	mysqlClient *database.Client
}

func NewService(
	mysqlClient *database.Client,
) *Service {
	return &Service{
		log:         zap.S().With("module", "services.system.service"),
		mysqlClient: mysqlClient,
	}
}

// CreateOperateLog 创建操作日志
func (s *Service) CreateOperateLog(
	ctx context.Context,
	log *models.OperateLogs,
) error {
	err := s.mysqlClient.Db().WithContext(ctx).
		Model(&models.OperateLogs{}).
		Create(log).
		Error
	if err != nil {
		return err
	}
	return nil
}
