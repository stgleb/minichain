package minichain

type Config struct {
	Main       MainConfig
	BlockChain BlockChainConfig
	Index      IndexConfig
	Http       HttpConfig
}

type MainConfig struct {
	LogLevel int
}

type BlockChainConfig struct {
	BlockSize    int
	TimeOut      int
	KeyMaxSize   int
	ValueMaxSize int
	DataFile     string
}

type IndexConfig struct {
	IndexType string
	IsOn      bool
}

type HttpConfig struct {
	ListenStr string
	Timeout   int64
}
