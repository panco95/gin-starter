package account

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"time"

	"lovebox/models"
	"lovebox/pkg/database"
	"lovebox/pkg/jwt"
	"lovebox/pkg/resp"
	"lovebox/pkg/utils"

	"github.com/afocus/captcha"
	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	log         *zap.SugaredLogger
	mysqlClient *database.Client
	redisClient *redis.Client
	cacheClient *redisCache.Cache
	jwt         *jwt.Jwt
}

func NewService(
	mysqlClient *database.Client,
	redisClient *redis.Client,
	cacheClient *redisCache.Cache,
	jwt *jwt.Jwt,
) *Service {
	return &Service{
		log:         zap.S().With("module", "services.account.service"),
		mysqlClient: mysqlClient,
		redisClient: redisClient,
		cacheClient: cacheClient,
		jwt:         jwt,
	}
}

// GetCaptcha 获取验证码
func (s *Service) GetCaptcha(
	ctx context.Context,
	prefix string,
) (*models.GetCaptchaRes, error) {
	key := uuid.New().String()
	cap := captcha.New()
	err := cap.AddFontFromBytes(utils.GetDefaultFont())
	if err != nil {
		s.log.Errorf("GetCaptcha cap.AddFontFromBytes err=%v", err)
		return nil, err
	}
	img, code := cap.Create(4, captcha.NUM)
	content := bytes.NewBuffer([]byte{})
	err = png.Encode(content, img)
	if err != nil {
		s.log.Errorf("GetCaptcha png.Encode err=%v", err)
		return nil, err
	}

	cacheKey := "captcha:" + prefix + ":" + key
	err = s.cacheClient.Set(&redisCache.Item{
		Ctx:   ctx,
		Key:   cacheKey,
		Value: code,
		TTL:   time.Minute * 5,
	})
	if err != nil {
		s.log.Errorf("GetCaptcha cacheClient.Set err=%v", err)
		return nil, err
	}

	return &models.GetCaptchaRes{
		Key:     key,
		Captcha: content.Bytes(),
	}, nil
}

// Login 账号登录
func (s *Service) Login(
	ctx context.Context,
	req *models.LoginReq,
	ip string,
) (string, *models.Account, error) {
	account, err := s.QueryAccount(ctx, &models.Account{
		Username: req.Username,
	})
	if err != nil {
		return "", nil, err
	}
	if account.ID == 0 {
		return "", nil, errors.New(resp.ACCOUNT_NOT_FOUND)
	}
	if account.Status == models.AccountStatusLock {
		return "", nil, errors.New(resp.ACCOUNT_LOCKED)
	}
	if utils.Md5(utils.Md5(req.Password)+account.PasswordSalt) != account.Password {
		return "", nil, errors.New(resp.ACCOUNT_PWD_ERROR)
	}

	token, err := s.jwt.BuildToken(
		account.ID,
		models.LoginExpired,
	)
	if err != nil {
		s.log.Errorf("Login jwt.BuildToken %v", err)
		return "", nil, errors.New(resp.SERVER_ERROR)
	}

	now := time.Now()
	err = s.mysqlClient.Db().
		Model(&account).
		Updates(models.Account{
			LastLoginTime: &now,
			LastLoginIp:   ip,
			LoginTimes:    account.LoginTimes + 1,
		}).
		Error
	if err != nil {
		s.log.Errorf("Login update account %v", err)
	}

	return token, account, nil
}

// Register 账号注册
func (s *Service) Register(
	ctx context.Context,
	req *models.RegisterReq,
	ip string,
) (string, *models.Account, error) {
	if utils.IsChinese(req.Username) {
		return "", nil, errors.New(resp.ACCOUNT_HAS_CHINESE)
	}
	account, err := s.QueryAccount(ctx, &models.Account{
		Username: req.Username,
	})
	if err != nil {
		return "", nil, err
	}
	if account.ID > 0 {
		return "", nil, errors.New(resp.ACCOUNT_EXISTS)
	}

	now := time.Now()
	account.Username = req.Username
	account.LastLoginTime = &now
	account.LoginTimes = 1
	account.PasswordSalt = utils.RandStr(6)
	account.Password = utils.Md5(utils.Md5(req.Password) + account.PasswordSalt)
	account.LastLoginIp = ip

	err = s.mysqlClient.Db().
		Model(&models.Account{}).
		Create(account).Error
	if err != nil {
		return "", nil, err
	}

	token, err := s.jwt.BuildToken(
		account.ID,
		models.LoginExpired,
	)
	if err != nil {
		s.log.Errorf("Register jwt.BuildToken %v", err)
		return "", nil, errors.New(resp.SERVER_ERROR)
	}

	go func() {
		err := s.mysqlClient.Db().
			Model(&models.AccountExtraInfo{}).
			Create(&models.AccountExtraInfo{
				AccoutnId: account.ID,
			}).Error
		if err != nil {
			s.log.Errorf("Register create AccountExtraInfo %v", err)
		}
	}()

	return token, account, nil
}

// Info 账号信息
func (s *Service) Info(
	ctx context.Context,
	accountId uint,
) (*models.InfoRes, error) {
	account := &models.Account{}
	account.ID = accountId
	account, err := s.QueryAccount(ctx, account)
	if err != nil {
		return nil, err
	}

	result := &models.InfoRes{
		Username: account.Username,
	}

	return result, nil
}
