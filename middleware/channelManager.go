package middleware

//被用来管理状态
type ChannelManagerStatus uint8

const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED ChannelManager = 0; //未初始化
	CHANNEL_MANAGER_STATUS_INITIALIZED ChannelManager = 1; //已初始化
	CHANNEL_MANAGER_STATUS_CLOSED ChannelManager = 2; //已关闭
)
