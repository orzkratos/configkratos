package configkratos

import (
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/stretchr/testify/require"
	"github.com/yyle88/must"
	"github.com/yyle88/neatjson/neatjsonm"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/rese"
)

func TestNewJsonSource(t *testing.T) {
	type Account struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
	}

	jsonSource := NewJsonSource(neatjsonm.B(&Account{
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

func TestNewJsonSource_Update(t *testing.T) {
	type DatabaseConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type AuthConfig struct {
		APIKey  string `json:"api_key"`
		Timeout int    `json:"timeout"`
	}

	type ServerConfig struct {
		AppName  string         `json:"app_name"`
		Version  string         `json:"version"`
		Database DatabaseConfig `json:"database"`
		Auth     AuthConfig     `json:"auth"`
		Port     int            `json:"port"`
	}

	type ConfigPortV2 struct {
		Port int `json:"port"`
	}

	type ConfigDatabaseV2 struct {
		Database DatabaseConfig `json:"database"`
	}

	type ConfigAuthV2 struct {
		Auth AuthConfig `json:"auth"`
	}

	jsonSource := NewJsonSource(neatjsonm.B(&ServerConfig{
		AppName: "my-app",
		Version: "1.0.0",
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "abc",
			Password: "xyz",
		},
		Auth: AuthConfig{
			APIKey:  "123",
			Timeout: 30,
		},
		Port: 8080,
	}))

	// 创建 Kratos 配置实例
	c := config.New(
		config.WithSource(
			jsonSource, // 通过 JSON 配置源加载
		),
	)
	defer rese.F0(c.Close)

	must.Done(c.Load())

	must.Done(c.Watch("app_name", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("version", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("port", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("database", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("database.port", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("auth", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))
	must.Done(c.Watch("auth.timeout", func(s string, value config.Value) {
		t.Log("watch-changes:", s, value)
	}))

	{ // 比较结果
		res := ServerConfig{}
		must.Done(c.Scan(&res))
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8080, res.Port)
		require.Equal(t, "localhost", res.Database.Host)
		require.Equal(t, 3306, res.Database.Port)
		require.Equal(t, "abc", res.Database.Username)
		require.Equal(t, "xyz", res.Database.Password)
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	must.Done(jsonSource.Update(neatjsonm.B(&ConfigPortV2{
		Port: 8081,
	})))
	time.Sleep(time.Millisecond * 100)

	{ // 比较结果
		res := ServerConfig{}
		must.Done(c.Scan(&res)) //只有重新 Scan 这里才能拿到新数据，根据前面的 Watch 观察到键的变动时就执行 Scan 也行，但无论如何都得手动调用 Scan 函数
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port)
		require.Equal(t, "localhost", res.Database.Host)
		require.Equal(t, 3306, res.Database.Port)
		require.Equal(t, "abc", res.Database.Username)
		require.Equal(t, "xyz", res.Database.Password)
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	must.Done(jsonSource.Update(neatjsonm.B(&ConfigDatabaseV2{
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     3307,
			Username: "aaa",
			Password: "xxx",
		},
	})))
	time.Sleep(time.Millisecond * 100)

	{ // 比较结果
		res := ServerConfig{}
		must.Done(c.Scan(&res)) //只有重新 Scan 这里才能拿到新数据
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port)
		require.Equal(t, "127.0.0.1", res.Database.Host)
		require.Equal(t, 3307, res.Database.Port)
		require.Equal(t, "aaa", res.Database.Username)
		require.Equal(t, "xxx", res.Database.Password)
		require.Equal(t, "123", res.Auth.APIKey)
		require.Equal(t, 30, res.Auth.Timeout)
	}

	must.Done(jsonSource.Update(neatjsonm.B(&ConfigAuthV2{
		Auth: AuthConfig{
			APIKey:  "uvw",
			Timeout: 36000,
		},
	})))
	time.Sleep(time.Millisecond * 100)

	{ // 比较结果
		res := ServerConfig{}
		must.Done(c.Scan(&res)) //只有重新 Scan 这里才能拿到新数据
		t.Log("config:", neatjsons.S(res))
		require.Equal(t, "my-app", res.AppName)
		require.Equal(t, "1.0.0", res.Version)
		require.Equal(t, 8081, res.Port)
		require.Equal(t, "127.0.0.1", res.Database.Host)
		require.Equal(t, 3307, res.Database.Port)
		require.Equal(t, "aaa", res.Database.Username)
		require.Equal(t, "xxx", res.Database.Password)
		require.Equal(t, "uvw", res.Auth.APIKey)
		require.Equal(t, 36000, res.Auth.Timeout)
	}
}
