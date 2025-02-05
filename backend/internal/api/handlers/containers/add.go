package containershandler

import (
	"app-pinger/backend/internal/api/utilapi"
	"app-pinger/backend/internal/entity"
	"app-pinger/pkg/contracts"
	"net/http"
	"time"
)

type ContainerAddResp struct {
	Text string `json:"msg"`
}

func (c *ContainersHandler) Add(ctx *utilapi.APIContext) {
	var req contracts.ContainerAddReq

	err := ctx.Decode(&req)
	if err != nil {
		ctx.Error("failed decode json", err)
		ctx.WriteFailure(http.StatusBadRequest, "invalid container")
		return
	}

	ctx.Debug("received request", "request", req)

	for _, r := range req.Containers {
		lastPing, err := time.Parse(time.DateTime, r.LastPing)
		if err != nil {
			ctx.Error("failed decode containers", err)
			ctx.WriteFailure(http.StatusBadRequest, "invalid request")
			return
		}

		container := entity.Container{
			IP:          r.IPAddress,
			IsReachable: r.IsReachable,
			LastPing:    lastPing,
			PacketLost:  r.PackerLost,
		}

		IP, err := c.containers.Add(ctx, container)
		if err != nil || IP != r.IPAddress {
			ctx.Error("failed to add container", err)
			ctx.WriteFailure(http.StatusInternalServerError, "internal error")
			return
		}
	}

	ctx.SuccessWithData(ContainerAddResp{"success"})
}
