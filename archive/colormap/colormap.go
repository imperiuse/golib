package colormap

import (
	"fmt"
	"strconv"

	"github.com/imperiuse/golib/archive/concat"
)

// CSN - ColorSchemeNumber
type CSN int // ColorSchemeNumber

//nolint
const (
	CsReset CSN = iota

	CsInfo // Настройка цветов сообщений (по типу сообщения)
	CsDebug
	CsWarning
	CsTest
	CsError
	CsFatalError
	CsPrint

	CsDb // Дополнительные цвета для сообщений (хранилище данных)
	CsDbOk
	CsDbFail

	CsRedis
	CsRedisOk
	CsRedisFail

	CsMemchd
	CsMemchdOk
	CsMemchdFail

	CsDatetime // Префикс логера(дата, время, место в коде)

	ColorschemeNext
)

// ColorSheme - it's element of Color Scheme Map
type ColorSheme []string

// CSM  - Color Scheme Map  - central unit of colormap package
type CSM map[CSN]ColorSheme

// Default Color Scheme Map
var defaultCSM = CSM{
	CsReset:      ColorSheme{CreateCS(ClrReset)},
	CsInfo:       ColorSheme{CreateCS(ClrFgGreen, ClrBold)},
	CsDebug:      ColorSheme{CreateCS(ClrFgYellow, ClrBold)},
	CsWarning:    ColorSheme{CreateCS(ClrFgYellow, ClrBold)},
	CsTest:       ColorSheme{CreateCS(ClrFgCyan, ClrBold)},
	CsError:      ColorSheme{CreateCS(ClrFgRed, ClrBold, ClrUnder)},
	CsFatalError: ColorSheme{CreateCS(ClrFgRed, ClrBold, ClrUnder)},
	CsPrint:      ColorSheme{CreateCS(ClrFgMagenta, ClrBold)},
	CsDb:         ColorSheme{CreateCS(ClrFgBlue, ClrBold), CreateCS(NewLine, ClrBgCyan, ClrBold)},
	CsRedis:      ColorSheme{CreateCS(ClrFgMagenta, ClrBold), CreateCS(NewLine, ClrBgMagenta, ClrBold)},
	CsMemchd:     ColorSheme{CreateCS(ClrFgGreen, ClrBold), CreateCS(NewLine, ClrBgGreen, ClrBold)},
	CsDatetime:   ColorSheme{CreateCS(ClrFgYellow)},
}

// CreateCS - create and return custom Color Scheme @see CS
func CreateCS(attrs ...interface{}) (s string) {
	s = "\x1b["
	for i, val := range attrs {
		if v, ok := val.(SchemeAttributes); ok {
			s = concat.Strings(s, strconv.FormatUint(uint64(v), 10))
			if i != len(attrs)-1 {
				s = concat.Strings(s, ";")
			}
		} else {
			s = concat.Strings(s, fmt.Sprint(val))
		}
	}
	s = concat.Strings(s, "m")
	return
}

// SchemeAttributes - number of Scheme attr
type SchemeAttributes int

//Шаблон для использования в современных командных оболочках и языках программирования таков:
//    \x1b[...m. Это ESCAPE-последовательность, где \x1b обозначает символ ESC (десятичный ASCII код 27),
// а вместо "..."  подставляются значения из таблицы, приведенной ниже, причем они могут комбинироваться,
// тогда нужно их перечислить через точку с запятой.

// An ANSI code: an escape character (ESC, ASCII value 27 \x1b in hex) followed by the appropriate code sequence:

//ON ATTRIBUTE
const (
	RepeatColor = ""
	NewLine     = "\n"
	Tab         = "\t"
	Tab4        = "\t\t\t\t"
	Tab8        = "\t\t\t\t\t\t\t\t"
	NLTab8      = "\n\t\t\t\t\t\t\t\t"

	ClrReset     SchemeAttributes = 0 // reset; clears all colors and styles (to white on black) // 0 нормальный режим
	ClrBold      SchemeAttributes = 1 // bold on                                                 // 1 жирный
	ClrItal      SchemeAttributes = 3 // italics on                                              // 3 курсив
	ClrUnder     SchemeAttributes = 4 // underline on                                            // 4 подчеркнутый
	ClrFlashing  SchemeAttributes = 5 // flashing                                                // 5 мигающий
	ClrInverse   SchemeAttributes = 7 // inverse on; reverses foreground & background colors     // 7 инвертированные цвета
	ClrInvisible SchemeAttributes = 8 // invisible                                               // 8 невидимый
	ClrStrike    SchemeAttributes = 9 // strikethrough on                                        // 9 зачеркнутый

	// OFF ATTRIBUTE
	ClrBoldOff      SchemeAttributes = 21 // bold off
	ClrItalOff      SchemeAttributes = 23 // italics off
	ClrUnderOff     SchemeAttributes = 24 // underline off
	ClrFlashingOff  SchemeAttributes = 25 // strikethrough off
	ClrInversOff    SchemeAttributes = 27 // inverse off
	ClrInvisibleOff SchemeAttributes = 28 // invisible off
	ClrStrikeOff    SchemeAttributes = 29 // strikethrough off

	// FOREGROUND COLOR                                                        //цвет  текста
	ClrFgBlack   SchemeAttributes = 30 // set foreground color to black               // 30	черный
	ClrFgRed     SchemeAttributes = 31 // set foreground color to red                 // 31	красный
	ClrFgGreen   SchemeAttributes = 32 // set foreground color to green               // 32	зеленый
	ClrFgYellow  SchemeAttributes = 33 // set foreground color to yellow              // 33	желтый
	ClrFgBlue    SchemeAttributes = 34 // set foreground color to blue                // 34	синий
	ClrFgMagenta SchemeAttributes = 35 // set foreground color to magenta (purple)    // 35	пурпурный
	ClrFgCyan    SchemeAttributes = 36 // set foreground color to cyan                // 36	голубой
	ClrFgWhite   SchemeAttributes = 37 // set foreground color to white               // 37	белый
	ClrFgDefault SchemeAttributes = 39 // set foreground color to default (white)     // 39 по-умолчанию (белый)

	// BACKGROUND COLOR                                                        // цвет фона
	ClrBgBlack   SchemeAttributes = 40 // set background color to black               // 40	черный
	ClrBgRed     SchemeAttributes = 41 // set background color to red                 // 41	красный
	ClrBgGreen   SchemeAttributes = 42 // set background color to green               // 42	зеленый
	ClrBgYellow  SchemeAttributes = 43 // set background color to yellow              // 43	желтый
	ClrBgBlue    SchemeAttributes = 44 // set background color to blue                // 44	синий
	ClrBgMagenta SchemeAttributes = 45 // set background color to magenta (purple)    // 45	пурпурный
	ClrBgCyan    SchemeAttributes = 46 // set background color to cyan                // 46	голубой
	ClrBgWhite   SchemeAttributes = 47 // set background color to white               // 47	белый
	ClrBgDefault SchemeAttributes = 49 // set background color to default (black)     // 49 по-умолчанию (черный)
)

func copyDefaultCSM(csm *CSM) {
	for key, value := range defaultCSM {
		(*csm)[key] = value
	}
}

// GetDefaultCSM - return Default Color Sheme Map @see CSM
func GetDefaultCSM() (cs CSM) {
	return CSMthemePicker("")
}

// CSMthemePicker - return Custom Color Sheme Map @see CSM
func CSMthemePicker(ThemeName string) (cs CSM) {
	cs = make(CSM)
	copyDefaultCSM(&cs)
	switch ThemeName {
	case "github":
		cs[CsDatetime] = ColorSheme{CreateCS(ClrFgDefault)}
		cs[CsInfo] = ColorSheme{CreateCS(ClrFgGreen, ClrBold)}
		cs[CsDebug] = ColorSheme{CreateCS(ClrFgYellow)}
		cs[CsWarning] = ColorSheme{CreateCS(ClrFgYellow, ClrBold)}
		cs[CsError] = ColorSheme{CreateCS(ClrFgRed, ClrBold)}
		cs[CsFatalError] = ColorSheme{CreateCS(ClrFgRed, ClrBold, ClrStrike)}
		cs[CsTest] = ColorSheme{CreateCS(ClrFgBlue)}
		cs[CsPrint] = ColorSheme{CreateCS(ClrFgMagenta, ClrBold)}
		cs[CsDb] = ColorSheme{CreateCS(ClrFgBlue, ClrBold), CreateCS(ClrBgCyan, ClrBold)}
		cs[CsDbOk] = ColorSheme{CreateCS(ClrFgYellow), CreateCS(ClrFgMagenta), CreateCS(ClrFgGreen, ClrBold), concat.Strings(NLTab8, CreateCS(ClrFgCyan, ClrBold)), NLTab8, CreateCS(ClrFgCyan, ClrBold)}
		cs[CsDbFail] = ColorSheme{CreateCS(ClrFgYellow), CreateCS(ClrFgMagenta), CreateCS(ClrFgRed, ClrBold), NLTab8, NLTab8, NLTab8, RepeatColor}
		cs[CsRedis] = ColorSheme{CreateCS(ClrFgMagenta, ClrBold), CreateCS(ClrFgBlue, ClrBold), NewLine, CreateCS(ClrBgMagenta, ClrBold)}
		cs[CsRedisOk] = ColorSheme{CreateCS(ClrFgYellow), CreateCS(ClrFgMagenta, ClrBold), CreateCS(ClrFgGreen, ClrBold), concat.Strings(NLTab8, CreateCS(ClrFgMagenta, ClrBold))}
		cs[CsRedisFail] = ColorSheme{CreateCS(ClrFgYellow), CreateCS(ClrFgRed, ClrBold), CreateCS(ClrFgRed, ClrBold), RepeatColor, NLTab8}
		cs[CsMemchd] = ColorSheme{CreateCS(ClrFgGreen, ClrBold), CreateCS(ClrFgBlue, ClrBold), NewLine, CreateCS(ClrBgGreen, ClrBold)}
		return
	default:
		return // copy of Deafult CSM
	}
}
