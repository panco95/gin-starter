package account

import (
	"context"

	"lovebox/models"
)

// QueryAccount 查询单个账号
func (s *Service) QueryAccount(
	ctx context.Context,
	account *models.Account,
) (*models.Account, error) {
	db := s.mysqlClient.Db().WithContext(ctx)
	nAccount := &models.Account{}
	result := db.
		Where("id = ? OR username = ?", account.ID, account.Username).
		First(nAccount)
	if result.Error != nil {
		return nAccount, result.Error
	}
	return nAccount, nil
}
