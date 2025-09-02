package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

// SupportedLanguage 支持的语言类型
type SupportedLanguage string

const (
	// LanguageEnglish 英语
	LanguageEnglish SupportedLanguage = "en"
	// LanguageChinese 中文
	LanguageChinese SupportedLanguage = "zh"
	// LanguageJapanese 日语
	LanguageJapanese SupportedLanguage = "ja"
	// LanguageKorean 韩语
	LanguageKorean SupportedLanguage = "ko"
)

// DefaultLanguage 默认语言
const DefaultLanguage = LanguageEnglish

// ContextKey 上下文键类型
type ContextKey string

const (
	// LanguageContextKey 语言上下文键
	LanguageContextKey ContextKey = "language"
)

// Translator 翻译器接口
type Translator interface {
	// Translate 翻译指定键的文本
	Translate(key string, args ...interface{}) string
	// TranslateWithLang 使用指定语言翻译文本
	TranslateWithLang(lang SupportedLanguage, key string, args ...interface{}) string
	// GetLanguage 获取当前语言
	GetLanguage() SupportedLanguage
	// SetLanguage 设置当前语言
	SetLanguage(lang SupportedLanguage)
	// LoadTranslations 加载翻译文件
	LoadTranslations(dir string) error
}

// translator 翻译器实现
type translator struct {
	mu           sync.RWMutex
	currentLang  SupportedLanguage
	translations map[SupportedLanguage]map[string]string
}

// NewTranslator 创建新的翻译器实例
// 参数:
//   - lang: 默认语言
// 返回值:
//   - Translator: 翻译器实例
func NewTranslator(lang SupportedLanguage) Translator {
	return &translator{
		currentLang:  lang,
		translations: make(map[SupportedLanguage]map[string]string),
	}
}

// Translate 翻译指定键的文本
// 参数:
//   - key: 翻译键
//   - args: 格式化参数
// 返回值:
//   - string: 翻译后的文本
func (t *translator) Translate(key string, args ...interface{}) string {
	return t.TranslateWithLang(t.GetLanguage(), key, args...)
}

// TranslateWithLang 使用指定语言翻译文本
// 参数:
//   - lang: 目标语言
//   - key: 翻译键
//   - args: 格式化参数
// 返回值:
//   - string: 翻译后的文本
func (t *translator) TranslateWithLang(lang SupportedLanguage, key string, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 尝试获取指定语言的翻译
	if langMap, exists := t.translations[lang]; exists {
		if text, found := langMap[key]; found {
			if len(args) > 0 {
				return fmt.Sprintf(text, args...)
			}
			return text
		}
	}

	// 回退到默认语言
	if lang != DefaultLanguage {
		if langMap, exists := t.translations[DefaultLanguage]; exists {
			if text, found := langMap[key]; found {
				if len(args) > 0 {
					return fmt.Sprintf(text, args...)
				}
				return text
			}
		}
	}

	// 如果都找不到，返回键本身
	return key
}

// GetLanguage 获取当前语言
// 返回值:
//   - SupportedLanguage: 当前语言
func (t *translator) GetLanguage() SupportedLanguage {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLang
}

// SetLanguage 设置当前语言
// 参数:
//   - lang: 要设置的语言
func (t *translator) SetLanguage(lang SupportedLanguage) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentLang = lang
}

// LoadTranslations 加载翻译文件
// 参数:
//   - dir: 翻译文件目录
// 返回值:
//   - error: 错误信息
func (t *translator) LoadTranslations(dir string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 支持的语言列表
	supportedLangs := []SupportedLanguage{
		LanguageEnglish,
		LanguageChinese,
		LanguageJapanese,
		LanguageKorean,
	}

	for _, lang := range supportedLangs {
		filePath := filepath.Join(dir, fmt.Sprintf("%s.json", string(lang)))
		if err := t.loadLanguageFile(lang, filePath); err != nil {
			// 如果不是默认语言文件，忽略错误
			if lang == DefaultLanguage {
				return fmt.Errorf("failed to load default language file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// loadLanguageFile 加载单个语言文件
// 参数:
//   - lang: 语言类型
//   - filePath: 文件路径
// 返回值:
//   - error: 错误信息
func (t *translator) loadLanguageFile(lang SupportedLanguage, filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 首先尝试解析为嵌套结构
	var nestedTranslations map[string]interface{}
	if err := json.Unmarshal(data, &nestedTranslations); err != nil {
		return fmt.Errorf("failed to parse translation file %s: %w", filePath, err)
	}

	if t.translations[lang] == nil {
		t.translations[lang] = make(map[string]string)
	}

	// 扁平化嵌套结构
	t.flattenTranslations("", nestedTranslations, t.translations[lang])

	return nil
}

// flattenTranslations 扁平化嵌套的翻译结构
// 参数:
//   - prefix: 键前缀
//   - nested: 嵌套的翻译数据
//   - flat: 扁平化后的翻译数据
func (t *translator) flattenTranslations(prefix string, nested map[string]interface{}, flat map[string]string) {
	for key, value := range nested {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			// 字符串值，直接添加
			flat[fullKey] = v
		case map[string]interface{}:
			// 嵌套对象，递归处理
			t.flattenTranslations(fullKey, v, flat)
		default:
			// 其他类型，转换为字符串
			flat[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

// GetLanguageFromContext 从上下文中获取语言
// 参数:
//   - ctx: 上下文
// 返回值:
//   - SupportedLanguage: 语言类型
func GetLanguageFromContext(ctx context.Context) SupportedLanguage {
	if lang, ok := ctx.Value(LanguageContextKey).(SupportedLanguage); ok {
		return lang
	}
	return DefaultLanguage
}

// SetLanguageToContext 将语言设置到上下文中
// 参数:
//   - ctx: 父上下文
//   - lang: 语言类型
// 返回值:
//   - context.Context: 新的上下文
func SetLanguageToContext(ctx context.Context, lang SupportedLanguage) context.Context {
	return context.WithValue(ctx, LanguageContextKey, lang)
}

// ParseAcceptLanguage 解析Accept-Language头
// 参数:
//   - acceptLang: Accept-Language头的值
// 返回值:
//   - SupportedLanguage: 解析出的语言
func ParseAcceptLanguage(acceptLang string) SupportedLanguage {
	if acceptLang == "" {
		return DefaultLanguage
	}

	// 简单解析Accept-Language头
	languages := strings.Split(acceptLang, ",")
	for _, lang := range languages {
		lang = strings.TrimSpace(lang)
		// 移除权重信息
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}

		// 匹配支持的语言
		switch {
		case strings.HasPrefix(lang, "zh"):
			return LanguageChinese
		case strings.HasPrefix(lang, "ja"):
			return LanguageJapanese
		case strings.HasPrefix(lang, "ko"):
			return LanguageKorean
		case strings.HasPrefix(lang, "en"):
			return LanguageEnglish
		}
	}

	return DefaultLanguage
}