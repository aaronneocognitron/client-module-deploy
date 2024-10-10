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
	Key          string `yaml:"Key"`
	Salt         string `yaml:"Salt"`
	CipherMethod string `yaml:"CipherMethod"`
}

type Db struct {
	Host     string `yaml:"Host"`
	Port     uint16 `yaml:"Port"`
	Name     string `yaml:"Name"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
}

type Fireblocks struct {
	ApiKey          string   `yaml:"ApiKey"`
	SecretPath      string   `yaml:"SecretPath"`
	VaultAccountIds []string `yaml:"VaultAccountIds"`
}

type Node struct {
	RPC string `yaml:"RPC"`

	// ton only
	ArchiveRpc *string `yaml:"ArchiveRPC"`

	// all expect solana
	ContractAddress *string `yaml:"ContractAddress"`

	OwnerPrivateKey *string `yaml:"OwnerPrivateKey,omitempty"`

	// everscale/venom only
	OwnerPublicKey *string `yaml:"OwnerPublicKey,omitempty"`

	// ton only
	OwnerWalletType *string `yaml:"OwnerWalletType,omitempty"`

	// solana only
	TokenProgramId          *string `yaml:"TokenProgramId"`
	TokenName               *string `yaml:"TokenName"`
	ClientProgramId         *string `yaml:"ClientProgramId"`
	ClientUserAddress       *string `yaml:"ClientUserAddress"`
	InitializerProgramId    *string `yaml:"InitializerProgramId"`
	RelayerProgramId        *string `yaml:"RelayerProgramId"`
	SystemRelayOwnerAddress *string `yaml:"SystemRelayOwnerAddress"`
	RelayOwnerAddress       *string `yaml:"RelayOwnerAddress"`

	MaxResendTries         int         `yaml:"MaxResendTries,omitempty"`
	MaxOutOfGasResendTries int         `yaml:"MaxOutOfGasResendTries,omitempty"`
	FeeMultiplierPercent   uint        `yaml:"FeeMultiplierPercent,omitempty"`
	Fireblocks             *Fireblocks `yaml:"Fireblocks,omitempty"`
}

type Utils struct {
	Encryption *Encryption `yaml:"Encryption"`
	Db         *Db         `yaml:"Db"`
}

type Config struct {
	Environment Environment `yaml:"Environment"`
	Utils       Utils       `yaml:"Utils"`
	Nodes       struct {
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

	if config.Environment.LogLevel == "" {
		config.Environment.LogLevel = "INFO"
	}

	// generate encryption
	if config.Utils.Encryption == nil {
		config.Utils.Encryption = &Encryption{}
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
			Name:     "asterizm-cs",
			User:     "asterizm-cs",
			Password: password,
		}
	}

	if len(config.Nodes.List) == 0 {
		return nil, errors.New("please, fill Nodes.List")
	}

	newList := make(map[string]Node, len(config.Nodes.List))

	for key, node := range config.Nodes.List {
		newList[key] = node

		if node.RPC == "" {
			return nil, fmt.Errorf("please, fill Nodes.List.%s.RPC", key)
		}
	}

	config.Nodes.List = newList
	return config, nil
}
