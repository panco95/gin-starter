package models

import "time"

const (
	LoginExpired = 7 * 24 * time.Hour
)

type Account struct {
	Model
	Username      string           `gorm:"column:username;not null;default:'';type:varchar(50);index:username" json:"username"` //用户名
	Nikcname      string           `gorm:"column:nickname;not null;default:'';type:varchar(50)" json:"nickname"`                //昵称
	Mobile        string           `gorm:"column:mobile;not null;default:'';type:varchar(50)" json:"mobile"`                    //手机号
	Avatar        string           `gorm:"column:avatar;not null;default:'';type:varchar(500)" json:"avatar"`                   //头像
	Gender        Gender           `gorm:"column:gender;not null;default:'';type:varchar(10)" json:"gender"`                    //性别
	Birth         string           `gorm:"column:birth;default:null;type:date" json:"birth"`                                    //生日
	Password      string           `gorm:"column:password;not null;default:'';type:varchar(200)" json:"password"`               //密码
	PasswordSalt  string           `gorm:"column:password_salt;not null;default:'';type:varchar(200)" json:"passwordSalt"`      //密码盐值
	Status        AccountStatus    `gorm:"column:status;not null;default:'normal';type:varchar(20)" json:"status"`              //状态
	LastLoginTime *time.Time       `gorm:"column:last_login_time;" json:"lastLoginTime"`                                        //最后登陆时间
	LastLoginIp   string           `gorm:"column:last_login_ip;not null;default:'';type:varchar(20)" json:"lastLoginIp"`        //最后登录IP
	LoginTimes    uint             `gorm:"column:login_times;not null;default:0;type:int(10)" json:"loginTimes"`                //登录次数
	ExtraInfo     AccountExtraInfo `gorm:"foreignKey:account_id"`
}

type AccountExtraInfo struct {
	Model           `json:"model"`
	AccoutnId       uint    `gorm:"column:account_id;not null;default:0" json:"userId"`                                                   //用户id
	Introduce       string  `gorm:"column:introduce;not null;default:'';type:varchar(500)" json:"introduce" binding:"max=500"`            //个人介绍
	ProfessionClass string  `gorm:"column:profession_class;not null;default:'';type:varchar(50)" json:"professionClass" binding:"max=50"` //职业类型
	Profession      string  `gorm:"column:profession;not null;default:'';type:varchar(50)" json:"profession" binding:"max=50"`            //职业
	Company         string  `gorm:"column:company;not null;default:'';type:varchar(50)" json:"company" binding:"max=50"`                  //公司
	Education       string  `gorm:"column:education;not null;default:'';type:varchar(50)" json:"education" binding:"max=50"`              //学历
	Country         string  `gorm:"column:country;not null;default:'';type:varchar(20)" json:"country" binding:"max=20"`                  //国家
	Province        string  `gorm:"column:province;not null;default:'';type:varchar(20)" json:"province" binding:"max=20"`                //省
	City            string  `gorm:"column:city;not null;default:'';type:varchar(20)" json:"city" binding:"max=20"`                        //市
	District        string  `gorm:"column:district;not null;default:'';type:varchar(20)" json:"district" binding:"max=20"`                //区
	Address         string  `gorm:"column:address;not null;default:'';type:varchar(200)" json:"address" binding:"max=200"`                //详细地址
	LookingFor      string  `gorm:"column:looking_for;not null;default:'';type:varchar(50)" json:"lookingFor" binding:"max=50"`           //交友目的
	SexTarget       string  `gorm:"column:sex_target;not null;default:'';type:varchar(20)" json:"sexTarget" binding:"max=20"`             //性取向
	HangOut         string  `gorm:"column:hang_out;not null;default:'';type:varchar(50)" json:"hangOut" binding:"max=50"`                 //经常出没
	Height          float32 `gorm:"column:height;not null;default:0;type:decimal(3,1)" json:"height" binding:"number"`                    //身高
	Weight          float32 `gorm:"column:weight;not null;default:0;type:decimal(3,1)" json:"weight" binding:"number"`                    //体重
	AnnualIncome    string  `gorm:"column:annual_income;not null;default:'';type:varchar(20)" json:"annualIncome" binding:"max=20"`       //年收入
	CarProperty     string  `gorm:"column:car_property;not null;default:'';type:varchar(100)" json:"carProperty" binding:"max=100"`       //车产
	HousePropetry   string  `gorm:"column:house_propetry;not null;default:'';type:varchar(100)" json:"housePropetry" binding:"max=100"`   //房产
	Labels          string  `gorm:"column:labels;type:text" json:"labels"`                                                                //个性标签
}

type AccountStatus string

var (
	AccountStatusNormal AccountStatus = "normal"
	AccountStatusLock   AccountStatus = "lock"
)

type Gender string

var (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type GetCaptchaReq struct {
	Type string `json:"type" binding:"required"`
}

type GetCaptchaRes struct {
	Key     string `json:"key"`
	Captcha []byte `json:"captcha"`
}

type LoginReq struct {
	Username string `form:"username" binding:"required,min=6,max=50"`
	Password string `form:"password" binding:"required,min=6,max=50"`
	// CaptchaKey string `form:"captchaKey" binding:"required"`
	// Captcha    string `form:"captcha" binding:"required"`
}

type RegisterReq struct {
	Username string `form:"username" binding:"required,min=6,max=50"`
	Password string `form:"password" binding:"required,min=6,max=50"`
	// CaptchaKey string `form:"captchaKey" binding:"required"`
	// Captcha    string `form:"captcha" binding:"required"`
}

type LoginOrRegisterRes struct {
	Token string `json:"token"`
}

type InfoRes struct {
	Username string `json:"username"`
}
