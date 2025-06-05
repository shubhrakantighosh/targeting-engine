package env

import (
	"context"
	"github.com/google/uuid"
	"main/constants"
)

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(constants.RequestID).(string)
	if !ok {
		return uuid.New().String()
	}

	return requestID
}
