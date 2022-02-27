package global

import (
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var DB *gorm.DB

var Config = viper.New()

var Cache *cache.Cache