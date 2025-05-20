package configkratos

import (
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsonm"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/rese"
	"gopkg.in/yaml.v3"
)

func TestNewYamlStatic(t *testing.T) {
	type ConfigType struct {
		Username string `yaml:"username"`
		Nickname string `yaml:"nickname"`
	}

	yamlSource := NewYamlStatic(rese.A1(yaml.Marshal(&ConfigType{
		Username: "abc",
		Nickname: "123",
	})))

	// 创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			yamlSource, // 通过 JSON 配置源加载
		),
	)
	defer rese.F0(c.Close)

	must.Done(c.Load())

	account := &ConfigType{}
	must.Done(c.Scan(account))
	t.Log("account:", string(rese.A1(yaml.Marshal(account))))
	require.Equal(t, "abc", account.Username)
	require.Equal(t, "123", account.Nickname)
}

func TestNewJsonStatic(t *testing.T) {
	type Account struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}

	jsonSource := NewJsonStatic(neatjsonm.B(&Account{
		Username: "abc",
		Nickname: "123",
	}))

	// 创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			jsonSource, // 通过 JSON 配置源加载
		),
	)
	defer rese.F0(c.Close)

	must.Done(c.Load())

	account := &Account{}
	must.Done(c.Scan(account))
	t.Log("account:", neatjsons.S(account))
	require.Equal(t, "abc", account.Username)
	require.Equal(t, "123", account.Nickname)
}
