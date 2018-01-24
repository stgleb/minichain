package minichain

type Config struct {
	Main       MainConfig
	BlockChain BlockChainConfig
	Http       HttpConfig
}

type MainConfig struct {
	LogLevel int
}

type BlockChainConfig struct {
	BlockSize int
	TimeOut   int
	DataDir   string
}

type HttpConfig struct {
	ListenStr    string
	ReadTimeout  int64
	WriteTimeout int64
}
