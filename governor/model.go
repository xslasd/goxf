package governor

type CodeInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CodeMap struct {
	Codes []CodeInfo `json:"codes"`
}

type APPInfo struct {
	APPID       string `json:"app_id"`
	ServiceName string `json:"name"`
	Version     string `json:"version"`
	BuildUser   string `json:"build_user"`
	BuildTime   int64  `json:"build_time"`
	StartTime   int64  `json:"start_time"`
}
