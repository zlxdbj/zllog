package zllog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ============================================================================
// 配置文件加载（独立于项目特定配置）
// ============================================================================

// ConfigLoader 配置加载器
// 支持多种配置来源，按优先级查找：
//   1. log.yaml（独立配置文件）
//   2. application.yaml（项目配置文件）
//   3. application_{ENV}.yaml（环境配置）
//   4. 默认配置
type ConfigLoader struct {
	// 配置文件查找目录（默认为当前目录）
	configDir string
	// 环境名称（dev/test/prod）
	envName string
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		configDir: "resource", // 默认从 resource 目录查找
		envName:   detectEnv(),
	}
}

// SetConfigDir 设置配置文件查找目录
func (l *ConfigLoader) SetConfigDir(dir string) {
	l.configDir = dir
}

// SetEnv 设置环境名称
func (l *ConfigLoader) SetEnv(env string) {
	l.envName = env
}

// LoadConfig 加载配置
// 按优先级查找配置文件，如果都找不到则使用默认配置
func (l *ConfigLoader) LoadConfig() *LogConfig {
	// 1. 尝试从 log.yaml 加载（独立配置文件）
	if config := l.loadFromLogYAML(); config != nil {
		return config
	}

	// 2. 尝试从 application.yaml 加载
	if config := l.loadFromAppYAML("application.yaml"); config != nil {
		return config
	}

	// 3. 尝试从 application_{ENV}.yaml 加载
	if l.envName != "" {
		appEnvFile := fmt.Sprintf("application_%s.yaml", l.envName)
		if config := l.loadFromAppYAML(appEnvFile); config != nil {
			return config
		}
	}

	// 4. 使用默认配置
	serviceName := detectServiceName()
	config := DefaultConfig(serviceName)
	adjustConfigByEnv(config)

	return config
}

// loadFromLogYAML 从独立的 log.yaml 加载配置
func (l *ConfigLoader) loadFromLogYAML() *LogConfig {
	configPath := filepath.Join(l.configDir, "log.yaml")

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	// 使用 viper 加载
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		// 文件存在但读取失败，返回 nil（使用默认配置）
		return nil
	}

	return l.parseLogConfig(v)
}

// loadFromAppYAML 从 application.yaml 加载 logger 配置
func (l *ConfigLoader) loadFromAppYAML(filename string) *LogConfig {
	configPath := filepath.Join(l.configDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	// 使用 viper 加载
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil
	}

	// 检查是否有 logger 配置项
	if !v.IsSet("logger") {
		return nil
	}

	// 从 logger 配置项读取
	return l.parseLoggerConfig(v)
}

// parseLogConfig 解析 log.yaml 配置（直接格式）
// log.yaml 格式：
//   service_name: my_service
//   env: dev
//   level: INFO
//   dir: ./logs
func (l *ConfigLoader) parseLogConfig(v *viper.Viper) *LogConfig {
	serviceName := detectServiceName()
	if v.IsSet("service_name") {
		serviceName = v.GetString("service_name")
	}

	config := DefaultConfig(serviceName)

	// 覆盖配置
	if v.IsSet("env") {
		config.Env = v.GetString("env")
	}
	if v.IsSet("level") {
		config.LogLevel = v.GetString("level")
	}
	if v.IsSet("dir") {
		config.LogDir = v.GetString("dir")
	}
	if v.IsSet("max_size") {
		config.MaxSize = v.GetInt("max_size")
	}
	if v.IsSet("max_backups") {
		config.MaxBackups = v.GetInt("max_backups")
	}
	if v.IsSet("max_age") {
		config.MaxAge = v.GetInt("max_age")
	}
	if v.IsSet("compress") {
		config.Compress = v.GetBool("compress")
	}
	if v.IsSet("daily_roll") {
		config.EnableDailyRoll = v.GetBool("daily_roll")
	}
	if v.IsSet("enable_console") {
		config.EnableConsole = v.GetBool("enable_console")
	}
	if v.IsSet("console_json") {
		config.ConsoleJSONFormat = v.GetBool("console_json")
	}

	// 根据环境调整配置
	adjustConfigByEnv(config)

	return config
}

// parseLoggerConfig 解析 application.yaml 中的 logger 配置
// application.yaml 格式：
//   logger:
//     level: INFO
//     dir: ./logs
func (l *ConfigLoader) parseLoggerConfig(v *viper.Viper) *LogConfig {
	serviceName := detectServiceName()

	// 尝试从 app.name 读取服务名
	if v.IsSet("app.name") {
		serviceName = v.GetString("app.name")
	}

	config := DefaultConfig(serviceName)

	// 从 logger 配置项读取
	if v.IsSet("logger.level") {
		config.LogLevel = v.GetString("logger.level")
	}
	if v.IsSet("logger.dir") {
		config.LogDir = v.GetString("logger.dir")
	}
	if v.IsSet("logger.max_size") {
		config.MaxSize = v.GetInt("logger.max_size")
	}
	if v.IsSet("logger.max_backups") {
		config.MaxBackups = v.GetInt("logger.max_backups")
	}
	if v.IsSet("logger.max_age") {
		config.MaxAge = v.GetInt("logger.max_age")
	}
	if v.IsSet("logger.compress") {
		config.Compress = v.GetBool("logger.compress")
	}
	if v.IsSet("logger.daily_roll") {
		config.EnableDailyRoll = v.GetBool("logger.daily_roll")
	}
	if v.IsSet("logger.enable_console") {
		config.EnableConsole = v.GetBool("logger.enable_console")
	}
	if v.IsSet("logger.console_json") {
		config.ConsoleJSONFormat = v.GetBool("logger.console_json")
	}

	// 根据环境调整配置
	adjustConfigByEnv(config)

	return config
}

// ============================================================================
// 辅助函数
// ============================================================================

// detectServiceName 自动检测服务名称
// 优先级: 环境变量 > 可执行文件名 > 当前目录名 > 默认值
func detectServiceName() string {
	// 方式1: 从环境变量读取
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		return name
	}
	if name := os.Getenv("APP_NAME"); name != "" {
		return name
	}

	// 方式2: 从可执行文件名获取
	if path, err := os.Executable(); err == nil {
		name := filepath.Base(path)
		// 去掉.exe后缀（Windows）
		name = strings.TrimSuffix(name, ".exe")
		if name != "" && name != "go" && name != "main" {
			return name
		}
	}

	// 方式3: 从当前目录名获取
	if dir, err := os.Getwd(); err == nil {
		name := filepath.Base(dir)
		if name != "" && name != "/" && name != "." {
			return name
		}
	}

	// 方式4: 使用默认值
	return "service"
}

// detectEnv 自动检测环境名称
func detectEnv() string {
	// 优先级: ENV > APP_ENV > GO_ENV > MODE > 默认 dev
	if env := os.Getenv("ENV"); env != "" {
		return env
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("GO_ENV"); env != "" {
		return env
	}
	if mode := os.Getenv("MODE"); mode != "" {
		return mode
	}
	return "dev"
}

// adjustConfigByEnv 根据环境智能调整配置
func adjustConfigByEnv(config *LogConfig) {
	// 如果没有手动配置环境，则自动检测
	if config.Env == "" || config.Env == "dev" {
		config.Env = detectEnv()
	}

	// 根据环境自动调整默认配置
	switch config.Env {
	case "prod", "production", "docker":
		if config.LogLevel == "" || config.LogLevel == "INFO" {
			config.LogLevel = "INFO"
		}
		if !config.EnableConsole {
			config.EnableConsole = false // 生产环境默认关闭控制台
		}
	case "test", "testing":
		if config.LogLevel == "" {
			config.LogLevel = "INFO"
		}
		config.EnableConsole = true
	case "dev", "development":
		if config.LogLevel == "" {
			config.LogLevel = "DEBUG"
		}
		config.EnableConsole = true
	}
}
