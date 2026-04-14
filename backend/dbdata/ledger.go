package dbdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// LedgerConfig 账本配置（存储当前活跃账本等信息）
type LedgerConfig struct {
	CurrentLedger   string            `json:"current_ledger"`    // 当前活跃账本名
	NameMap         map[string]string `json:"name_map"`          // filename -> 显示名称

}

const ledgerConfigFile = "ledger_config.json"

// GetLedgerDir 获取账本目录（~/.piggy-accounting/ledgers/）
func GetLedgerDir() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(dataDir, "ledgers")
	return dir, os.MkdirAll(dir, 0755)
}

// LoadLedgerConfig 加载账本配置
func LoadLedgerConfig() (*LedgerConfig, error) {
	configPath, err := getLedgerConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &LedgerConfig{CurrentLedger: "", NameMap: make(map[string]string)}, nil
		}
		return nil, err
	}

	var config LedgerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return &LedgerConfig{CurrentLedger: "", NameMap: make(map[string]string)}, nil
	}
	if config.NameMap == nil {
		config.NameMap = make(map[string]string)
	}
	return &config, nil
}

// SaveLedgerConfig 保存账本配置
func SaveLedgerConfig(config *LedgerConfig) error {
	configPath, err := getLedgerConfigPath()
	if err != nil {
		return err
	}
	if config.NameMap == nil {
		config.NameMap = make(map[string]string)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// ==================== 路径计算（供 dbdata 和 service 共用） ====================

// ResolveLedgerFilename 根据账本显示名解析出文件名
// 优先查 NameMap 精确匹配，其次直接拼接，最后用 hash 兜底
func ResolveLedgerFilename(name string, config *LedgerConfig) string {
	return resolveLedgerFilename(name, config)
}

func resolveLedgerFilename(name string, config *LedgerConfig) string {
	if config != nil {
		for filename, display := range config.NameMap {
			if display == name {
				return filename
			}
		}
	}
	return ledgerNameToFilename(name)
}

// LedgerNameToFilename 将账本显示名转为安全的文件名
func LedgerNameToFilename(name string) string {
	return ledgerNameToFilename(name)
}

// ledgerNameToFilename 将账本显示名转为安全的文件名
// 不含非法字符时直接用原名；含非法字符时用安全字符+hash后缀
func ledgerNameToFilename(name string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "默认账本.db"
	}

	hasInvalidChars := strings.ContainsAny(trimmedName, `\/:*?"<>|`)
	if !hasInvalidChars {
		return trimmedName + ".db"
	}

	// 含非法字符时转换
	safe := strings.Map(func(r rune) rune {
		if r == ' ' || r == '_' || r == '-' {
			return '_'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, trimmedName)
	if safe == "" {
		safe = "默认账本"
	}
	if len(safe) > 30 {
		safe = safe[:30]
	}
	safe = strings.TrimRight(safe, "_")
	if safe == "" {
		safe = "默认账本"
	}
	hash := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%s_%s.db", safe, hex.EncodeToString(hash[:8]))
}

// ==================== 内部辅助函数 ====================

func getLedgerConfigPath() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, ledgerConfigFile), nil
}

func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".piggy-accounting"), nil
}


