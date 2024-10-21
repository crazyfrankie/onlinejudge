package zapx

import (
	"go.uber.org/zap/zapcore"
)

type CustomCore struct {
	zapcore.Core
}

func NewCustomCore(core zapcore.Core) *CustomCore {
	return &CustomCore{
		Core: core,
	}
}

func (z *CustomCore) Write(en zapcore.Entry, fields []zapcore.Field) error {
	for _, fd := range fields {
		if fd.Key == "phone" {
			phone := fd.String
			fd.String = phone[:3] + "****" + phone[7:]
		}
	}

	return z.Core.Write(en, fields)
}
