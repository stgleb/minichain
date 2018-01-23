package minichain

type Config struct {
	Main       MainConfig
	BlockChain BlockChainConfig
}

type MainConfig struct {
	ListenStr string
	LogLevel  int
}

type BlockChainConfig struct {
	BlockSize int
	TimeOut   int
	DataDir   string
}
