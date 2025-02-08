package containershandler

import (
	"app-pinger/backend/internal/api/utilapi"
	"net/http"
	"time"
)

type ContainersResp struct {
	IPAddress   string `json:"ip_address"`
	IsReachable bool   `json:"is_reachable"`
	LastPing    string `json:"last_ping"`
}

func (c *ContainersHandler) GetAll(ctx *utilapi.APIContext) {
	containers, err := c.containers.GetAll(ctx)
	if err != nil {
		ctx.Error("failed to get all containers", err)
		ctx.WriteFailure(http.StatusInternalServerError, "internal error")
		return
	}

	data := make([]ContainersResp, len(containers))

	for i, container := range containers {

		data[i] = ContainersResp{
			IPAddress:   container.IP,
			IsReachable: container.IsReachable,
			LastPing:    container.LastPing.Format(time.DateTime),
		}
	}

	ctx.SuccessWithData(data)
}
