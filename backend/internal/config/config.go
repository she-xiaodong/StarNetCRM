package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	Neo4j    Neo4jConfig    `yaml:"neo4j"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	WeCom    WeComConfig    `yaml:"wecom"`
	Deploy   DeployConfig   `yaml:"deploy"`
	Security SecurityConfig `yaml:"security"`
}

type ServerConfig struct {
	Mode string `yaml:"mode"` // debug / release
	Port int    `yaml:"port"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
}

func (m MySQLConfig) DSN() string {
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + itoa(m.Port) + ")/" + m.Database + "?charset=" + m.Charset + "&parseTime=True&loc=Local"
}

type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"` // 默认 neo4j
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (r RedisConfig) Addr() string {
	return r.Host + ":" + itoa(r.Port)
}

type JWTConfig struct {
	Secret    string `yaml:"secret"`
	ExpireH   int    `yaml:"expire_hours"` // token过期时间（小时）
}

type WeComConfig struct {
	CorpID          string `yaml:"corp_id"`
	AgentID         string `yaml:"agent_id"`
	Secret          string `yaml:"secret"`
	RedirectURI     string `yaml:"redirect_uri"`
	Token           string `yaml:"token"`            // 回调验证Token
	EncodingAESKey  string `yaml:"encoding_aes_key"`  // 回调消息加解密Key
}

type DeployConfig struct {
	Mode string `yaml:"mode"` // saas / standalone
}

type SecurityConfig struct {
	AESKey       string `yaml:"aes_key"`        // AES-256加密密钥
	LogRetention int    `yaml:"log_retention"`   // 审计日志保留天数
}

var AppConfig *Config

// Load 从YAML文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		MySQL: MySQLConfig{
			Charset: "utf8mb4",
		},
		Neo4j: Neo4jConfig{
			Database: "neo4j",
		},
		Redis: RedisConfig{
			DB: 0,
		},
		JWT: JWTConfig{
			ExpireH: 24,
		},
		Deploy: DeployConfig{
			Mode: "saas",
		},
		Security: SecurityConfig{
			LogRetention: 180,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 环境变量覆盖
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = atoi(v)
	}
	if v := os.Getenv("DEPLOY_MODE"); v != "" {
		cfg.Deploy.Mode = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}

	AppConfig = cfg
	return cfg, nil
}

func itoa(i int) string {
	return formatInt(i)
}

func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var buf [20]byte
	p := len(buf)
	for i > 0 {
		p--
		buf[p] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		p--
		buf[p] = '-'
	}
	return string(buf[p:])
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
