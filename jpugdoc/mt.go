package jpugdoc

import (
	"fmt"
	"io"

	"github.com/jwalton/gchalk"
	"github.com/noborus/go-textra"
)

// MT はoriginを機械翻訳し出力します。
func MT(w io.Writer, origin string) error {
	// 翻訳タイプ
	var translates = []string{
		Config.APIAutoTranslateType, // PostgreSQLマニュアル翻訳
		textra.GENERAL_EN_JA,        // 一般翻訳
	}

	cli, err := newTextra(Config)
	if err != nil {
		return fmt.Errorf("textra: %s", err)
	}

	for _, apiType := range translates {
		ja, err := cli.Translate(apiType, origin)
		if err != nil {
			return fmt.Errorf("textra: %s", err)
		}
		if _, err := fmt.Fprintf(w, "%s: %s\n", gchalk.Green(apiType), ja); err != nil {
			return err
		}
	}
	return nil
}
