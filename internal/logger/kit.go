package logger

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/rs/zerolog"
)

type ZerologKitBridge struct {
	zl zerolog.Logger
}

var _ log.Logger = (*ZerologKitBridge)(nil)

type logPair struct {
	key string
	val any
}

func NewZerologKitBridge(zl zerolog.Logger) *ZerologKitBridge {
	return &ZerologKitBridge{zl: zl}
}

func (b *ZerologKitBridge) Log(keyVals ...any) error {
	lvl := zerolog.InfoLevel
	pairs := make([]logPair, 0, len(keyVals)/2+1)
	for i := 0; i < len(keyVals); i += 2 {
		key, ok := keyVals[i].(string)
		if !ok || i+1 >= len(keyVals) {
			continue // Go Kit keys must be strings
		}
		val := keyVals[i+1]
		if key == level.Key() {
			switch fmt.Sprint(val) {
			case level.DebugValue().String():
				lvl = zerolog.DebugLevel
			case level.InfoValue().String():
				lvl = zerolog.InfoLevel
			case level.WarnValue().String():
				lvl = zerolog.WarnLevel
			case level.ErrorValue().String():
				lvl = zerolog.ErrorLevel
			}
			continue
		}
		pairs = append(pairs, logPair{key: key, val: val})
	}
	evt := b.zl.WithLevel(lvl)
	for _, pair := range pairs {
		evt = evt.Interface(pair.key, pair.val)
	}
	evt.Send()
	return nil
}
