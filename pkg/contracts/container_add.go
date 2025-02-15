package contracts

import "unicode/utf8"

type PingData struct {
	IPAddress   string `json:"ip_address"`
	IsReachable bool   `json:"is_reachable"`
	LastPing    string `json:"last_ping"`
}

type ContainerAddReq struct {
	Containers []PingData `json:"containers"`
}

func (req *ContainerAddReq) IsValid() bool {
	for _, r := range req.Containers {
		if utf8.RuneCountInString(r.IPAddress) > 0 && utf8.RuneCountInString(r.LastPing) > 0 {
			return true
		}
	}

	return false
}
