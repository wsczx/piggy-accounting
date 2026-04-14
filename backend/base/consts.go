package base

import "log/slog"

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12) // 自定义比 Error 更高的级别
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

// ANSI 颜色代码
const (
	ColorReset = "\033[0m" // 重置/还原
	ColorBold  = "\033[1m" // 加粗/高亮
	ColorDim   = "\033[2m" // 暗色/调淡

	// 基础颜色
	ColorRed          = "\033[31m"       // 红色
	ColorGreen        = "\033[32m"       // 绿色
	ColorYellow       = "\033[33m"       // 黄色
	ColorBlue         = "\033[34m"       // 蓝色
	ColorMagenta      = "\033[35m"       // 深紫色
	ColorLightMagenta = "\033[95m"       // 高亮度深紫色
	ColorLavender     = "\033[38;5;183m" // 浅紫色
	ColorCyan         = "\033[36m"       // 青色
	ColorWhite        = "\033[37m"       // 白色

	// 组合样式
	// styleTitle  = colorBold + colorMagenta // 加粗紫
	// styleTitle = colorBold + colorLavender // 亮紫色
)

// 业务模块：用户相关
const (
	MsgMessage      = "MESSAGE"
	MsgUserLogin    = "USER_LOGIN"
	MsgUserLogout   = "USER_LOGOUT"
	MsgUserRegister = "USER_REGISTER"
)

// 系统模块：数据库/网络相关
const (
	MsgDBConnectError = "DATABASE_CONNECT_ERROR"
	MsgDBQueryTimeout = "DATABASE_QUERY_TIMEOUT"
	MsgConfigLoaded   = "CONFIG_LOADED_SUCCESS"
)

// 协议/网关相关
const (
	MsgReqReceived = "REQUEST_RECEIVED"
	MsgResSent     = "RESPONSE_SENT"
)
