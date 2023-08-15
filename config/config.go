package config

import (
	"asterizm/builder/utils"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Environment struct {
	LogLevel string `yaml:"LogLevel"`
}

type Encryption struct {
	EncryptPayload bool   `yaml:"EncryptPayload"`
	Key            string `yaml:"Key"`
	Salt           string `yaml:"Salt"`
	CipherMethod   string `yaml:"CipherMethod"`
}

type Db struct {
	Host     string `yaml:"Host"`
	Port     uint16 `yaml:"Port"`
	Name     string `yaml:"Name"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
}

type AsterizmTranslator struct {
	Host   string `yaml:"Host"`
	ApiKey string `yaml:"ApiKey"`
}

type Node struct {
	RPC                  string `yaml:"RPC"`
	ChainId              string `yaml:"ChainId"`
	ContractAddress      string `yaml:"ContractAddress"`
	OwnerAddress         string `yaml:"OwnerAddress"`
	OwnerPublicKey       string `yaml:"OwnerPublicKey,omitempty"`
	OwnerPrivateKey      string `yaml:"OwnerPrivateKey"`
	MaxResendTries       int    `yaml:"MaxResendTries,omitempty"`
	FeeMultiplierPercent uint   `yaml:"FeeMultiplierPercent,omitempty"`
}

type Utils struct {
	Encryption         *Encryption        `yaml:"Encryption"`
	Db                 *Db                `yaml:"Db"`
	AsterizmTranslator AsterizmTranslator `yaml:"AsterizmTranslator"`
}

type Config struct {
	Environment Environment `yaml:"Environment"`
	Utils       Utils       `yaml:"Utils"`
	Nodes       struct {
		ForceOrder    bool            `yaml:"ForceOrder"`
		PayloadStruct []string        `yaml:"PayloadStruct"`
		List          map[string]Node `yaml:"List"`
	} `yaml:"Nodes"`
}

func ParseAndRefreshConfig(dockerDbHost, configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling yaml: %w", err)
	}

	if config.Utils.AsterizmTranslator.Host == "" {
		return nil, errors.New("please, fill Utils.AsterizmTranslator.Host")
	}

	if config.Utils.AsterizmTranslator.ApiKey == "" {
		return nil, errors.New("please, fill Utils.AsterizmTranslator.ApiKey")
	}

	if config.Environment.LogLevel == "" {
		config.Environment.LogLevel = "INFO"
	}

	// generate encryption
	if config.Utils.Encryption == nil {
		config.Utils.Encryption = &Encryption{
			EncryptPayload: true,
		}
	}

	if config.Utils.Encryption.Key == "" {
		key, err := utils.GenerateEncryptionString(48)
		if err != nil {
			return nil, fmt.Errorf("generate encryption key: %w", err)
		}
		config.Utils.Encryption.Key = key
	}

	if config.Utils.Encryption.Salt == "" {
		salt, err := utils.GenerateEncryptionString(48)
		if err != nil {
			return nil, fmt.Errorf("generate encryption salt: %w", err)
		}
		config.Utils.Encryption.Salt = salt
	}

	if config.Utils.Encryption.CipherMethod == "" {
		config.Utils.Encryption.CipherMethod = "AES-256-CBC"
	}

	// generate db
	if config.Utils.Db == nil {
		password, err := utils.GeneratePassword(32)
		if err != nil {
			return nil, fmt.Errorf("generate db password: %w", err)
		}
		config.Utils.Db = &Db{
			Host:     dockerDbHost,
			Port:     5432,
			Name:     "asterizm",
			User:     "asterizm",
			Password: password,
		}
	}

	if len(config.Nodes.List) == 0 {
		return nil, errors.New("please, fill Nodes.List")
	}

	encryptor := utils.NewEncryptor(config.Utils.Encryption.Key, config.Utils.Encryption.Salt, config.Utils.Encryption.CipherMethod)
	newList := make(map[string]Node, len(config.Nodes.List))

	for key, node := range config.Nodes.List {
		if node.RPC == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.RPC", key)
		}

		if node.ChainId == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.ChainId", key)
		}

		if node.ContractAddress == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.ContractAddress", key)
		}

		if node.OwnerAddress == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.OwnerAddress", key)
		}

		if node.OwnerPrivateKey == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.OwnerPrivateKey", key)
		}

		if _, err := encryptor.Decrypt([]byte(node.OwnerAddress), "", ""); err != nil {
			// owner address haven't encrypted before
			encrypted, err := encryptor.Encrypt([]byte(node.OwnerAddress), "", "")
			if err != nil {
				return nil, fmt.Errorf("encrypt owner address: %w", err)
			}
			node.OwnerAddress = string(encrypted)
		}

		if _, err := encryptor.Decrypt([]byte(node.OwnerPrivateKey), "", ""); err != nil {
			// owner private key haven't encrypted before
			encrypted, err := encryptor.Encrypt([]byte(node.OwnerPrivateKey), "", "")
			if err != nil {
				return nil, fmt.Errorf("encrypt owner private key: %w", err)
			}
			node.OwnerPrivateKey = string(encrypted)
		}

		if node.OwnerPublicKey != "" {
			if _, err := encryptor.Decrypt([]byte(node.OwnerPublicKey), "", ""); err != nil {
				// owner public key haven't encrypted before
				encrypted, err := encryptor.Encrypt([]byte(node.OwnerPublicKey), "", "")
				if err != nil {
					return nil, fmt.Errorf("encrypt owner public key: %w", err)
				}
				node.OwnerPublicKey = string(encrypted)
			}
		}

		newList[key] = node
	}

	config.Nodes.List = newList
	return config, nil
}
