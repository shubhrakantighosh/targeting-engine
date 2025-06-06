package util

import (
	"context"
	"fmt"
	"main/pkg/env"
)

func LogPrefix(ctx context.Context, funcName string) string {
	return fmt.Sprintf("RequestID : %s func : %s ", env.GetRequestID(ctx), funcName)
}
