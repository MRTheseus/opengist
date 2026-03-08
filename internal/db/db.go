package db

import (
	"regexp"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex,size:191"`
	Password  string
	CreatedAt int64
}

type Visibility int

const (
	PublicVisibility  Visibility = 0
	PrivateVisibility Visibility = 1
)

func (v Visibility) String() string {
	if v == PrivateVisibility {
		return "private"
	}
	return "public"
}

type Gist struct {
	ID         uint       `gorm:"primaryKey"`
	Title      string     `gorm:"uniqueIndex,size:191"`
	Filename   string
	Content    string
	Visibility Visibility `gorm:"default:0"`
	UserID     uint
	CreatedAt  int64
	UpdatedAt  int64
}

type Revision struct {
	ID         uint `gorm:"primaryKey"`
	GistID     uint `gorm:"index"`
	Version    int
	Title      string
	Filename   string
	Content    string
	Visibility Visibility
	CreatedAt  int64
}

var titleRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func IsValidTitle(title string) bool {
	return titleRegex.MatchString(title)
}

func Setup(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	log.Info().Msg("Database connection established")

	if err = DB.AutoMigrate(&User{}, &Gist{}, &Revision{}); err != nil {
		return err
	}

	if err = initAdminUser(); err != nil {
		return err
	}

	return nil
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func initAdminUser() error {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := &User{
			Username:  "admin",
			Password:  string(hashedPassword),
			CreatedAt: time.Now().Unix(),
		}
		if err := DB.Create(admin).Error; err != nil {
			return err
		}
		log.Info().Msg("Default admin user created (password: 123456)")
	}
	return nil
}

func GetUserByUsername(username string) (*User, error) {
	user := new(User)
	err := DB.Where("username = ?", username).First(&user).Error
	return user, err
}

func GetUserById(id uint) (*User, error) {
	user := new(User)
	err := DB.Where("id = ?", id).First(&user).Error
	return user, err
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) Update() error {
	return DB.Save(u).Error
}

func TitleExists(title string, excludeId uint) (bool, error) {
	var count int64
	query := DB.Model(&Gist{}).Where("title = ?", title)
	if excludeId > 0 {
		query = query.Where("id != ?", excludeId)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (g *Gist) Create() error {
	g.CreatedAt = time.Now().Unix()
	g.UpdatedAt = g.CreatedAt
	return DB.Create(g).Error
}

func (g *Gist) Update() error {
	g.UpdatedAt = time.Now().Unix()
	return DB.Save(g).Error
}

func (g *Gist) Delete() error {
	return DB.Delete(g).Error
}

func GetGistByTitle(title string) (*Gist, error) {
	gist := new(Gist)
	err := DB.Where("title = ?", title).First(&gist).Error
	return gist, err
}

func GetGistById(id uint) (*Gist, error) {
	gist := new(Gist)
	err := DB.Where("id = ?", id).First(&gist).Error
	return gist, err
}

func GetAllGists(offset int, sort string, order string) ([]*Gist, error) {
	var gists []*Gist
	err := DB.Limit(11).
		Offset(offset * 10).
		Order(sort + "_at " + order).
		Find(&gists).Error
	return gists, err
}

func SearchGists(query string, offset int, sort string, order string) ([]*Gist, error) {
	var gists []*Gist
	err := DB.Where("title LIKE ?", "%"+query+"%").
		Limit(11).
		Offset(offset * 10).
		Order(sort + "_at " + order).
		Find(&gists).Error
	return gists, err
}

func CountAllGists() (int64, error) {
	var count int64
	err := DB.Model(&Gist{}).Count(&count).Error
	return count, err
}

func CountSearchGists(query string) (int64, error) {
	var count int64
	err := DB.Model(&Gist{}).Where("title LIKE ?", "%"+query+"%").Count(&count).Error
	return count, err
}

func (r *Revision) Create() error {
	r.CreatedAt = time.Now().Unix()
	return DB.Create(r).Error
}

func GetNextVersion(gistId uint) (int, error) {
	var maxVersion int
	err := DB.Model(&Revision{}).
		Where("gist_id = ?", gistId).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	return maxVersion + 1, err
}

func GetRevisionsByGistId(gistId uint) ([]*Revision, error) {
	var revisions []*Revision
	err := DB.Where("gist_id = ?", gistId).
		Order("version DESC").
		Find(&revisions).Error
	return revisions, err
}

func GetRevisionByGistIdAndVersion(gistId uint, version int) (*Revision, error) {
	revision := new(Revision)
	err := DB.Where("gist_id = ? AND version = ?", gistId, version).First(&revision).Error
	return revision, err
}

func DeleteRevisionsByGistId(gistId uint) error {
	return DB.Where("gist_id = ?", gistId).Delete(&Revision{}).Error
}
