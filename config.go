package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
)

var defaultConfigDirectory string

func init() {
	cu, err := user.Current()
	if err != nil {
		panic(err)
	}
	defaultConfigDirectory = filepath.Join(cu.HomeDir, ".config", "ylc")
}

type Config struct {
	dir string
}

func NewConfig() *Config {
	return &Config{
		dir: defaultConfigDirectory, // TODO: to be configurable
	}
}

func (c *Config) TokenStore() *TokenStore {
	return &TokenStore{
		config:   c,
		fileName: "token.json",
	}
}

func (c *Config) writeFile(fileName string, data []byte, perm os.FileMode) error {
	// ensure directory exists
	err := os.MkdirAll(c.dir, 0755)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(c.dir, fileName), data, perm)
}

func (c *Config) readFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(c.dir, fileName))
}

type TokenStore struct {
	config   *Config
	fileName string
}

func (ts *TokenStore) Save(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return ts.config.writeFile(ts.fileName, data, 0600)
}

func (ts *TokenStore) Load() (*oauth2.Token, error) {
	data, err := ts.config.readFile(ts.fileName)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(data, &token)
	return &token, err
}
