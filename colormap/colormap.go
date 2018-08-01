package colormap

import (
	"fmt"
	"github.com/golang_lib/concat"
	"strconv"
)

type CSN int // ColorSchemeNumber

const (
	CS_RESET CSN = iota
	// Настройка цветов сообщений (по типу сообщения)
	CS_INFO
	CS_DEBUG
	CS_WARNING
	CS_TEST
	CS_ERROR
	CS_FATAL_ERROR
	CS_PRINT

	// Дополнительные цвета для сообщений (хранилище данных)
	CS_DB
	CS_DB_OK
	CS_DB_FAIL

	CS_REDIS
	CS_REDIS_OK
	CS_REDIS_FAIL

	CS_MEMCHD
	CS_MEMCHD_OK
	CS_MEMCHD_FAIL

	// Префикс логера(дата, время, место в коде)
	CS_DATETIME

	COLORSCHEME_NEXT
)

type ColorSheme []string
type CSM map[CSN]ColorSheme // ColorSchemeMap

var DefaultCSM = CSM{
	CS_RESET:       ColorSheme{CreateCS(CLR_RESET)},
	CS_INFO:        ColorSheme{CreateCS(CLR_FG_GREEN, CLR_BOLD)},
	CS_DEBUG:       ColorSheme{CreateCS(CLR_FG_YELLOW, CLR_BOLD)},
	CS_WARNING:     ColorSheme{CreateCS(CLR_FG_YELLOW, CLR_BOLD)},
	CS_TEST:        ColorSheme{CreateCS(CLR_FG_CYAN, CLR_BOLD)},
	CS_ERROR:       ColorSheme{CreateCS(CLR_FG_RED, CLR_BOLD, CLR_UNDER)},
	CS_FATAL_ERROR: ColorSheme{CreateCS(CLR_FG_RED, CLR_BOLD, CLR_UNDER)},
	CS_PRINT:       ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD)},
	CS_DB:          ColorSheme{CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_CYAN, CLR_BOLD)},
	CS_REDIS:       ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_MAGENTA, CLR_BOLD)},
	CS_MEMCHD:      ColorSheme{CreateCS(CLR_FG_GREEN, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_GREEN, CLR_BOLD)},
	CS_DATETIME:    ColorSheme{CreateCS(CLR_FG_YELLOW)},
}

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

type SchemeAttributes int

//Шаблон для использования в современных командных оболочках и языках программирования таков:
//    \x1b[...m. Это ESCAPE-последовательность, где \x1b обозначает символ ESC (десятичный ASCII код 27),
// а вместо "..."  подставляются значения из таблицы, приведенной ниже, причем они могут комбинироваться,
// тогда нужно их перечислить через точку с запятой.

// An ANSI code: an escape character (ESC, ASCII value 27 \x1b in hex) followed by the appropriate code sequence:

//ON ATTRIBUTE
const (
	NEW_LINE = "\n"
	TAB      = "\t"

	CLR_RESET     SchemeAttributes = 0 // reset; clears all colors and styles (to white on black) // 0 нормальный режим
	CLR_BOLD      SchemeAttributes = 1 // bold on                                                 // 1 жирный
	CLR_ITAL      SchemeAttributes = 3 // italics on                                              // 3 курсив
	CLR_UNDER     SchemeAttributes = 4 // underline on                                            // 4 подчеркнутый
	CLR_FLASHING  SchemeAttributes = 5 // flashing                                                // 5 мигающий
	CLR_INVERSE   SchemeAttributes = 7 // inverse on; reverses foreground & background colors     // 7 инвертированные цвета
	CLR_INVISIBLE SchemeAttributes = 8 // invisible                                               // 8 невидимый
	CLR_STRIKE    SchemeAttributes = 9 // strikethrough on                                        // 9 зачеркнутый

	// OFF ATTRIBUTE
	CLR_BOLD_OFF      SchemeAttributes = 21 // bold off
	CLR_ITAL_OFF      SchemeAttributes = 23 // italics off
	CLR_UNDER_OFF     SchemeAttributes = 24 // underline off
	CLR_FLASHING_OFF  SchemeAttributes = 25 // strikethrough off
	CLR_INVERS_OFF    SchemeAttributes = 27 // inverse off
	CLR_INVISIBLE_OFF SchemeAttributes = 28 // invisible off
	CLR_STRIKE_OFF    SchemeAttributes = 29 // strikethrough off

	// FOREGROUND COLOR                                                        //цвет  текста
	CLR_FG_BLACK   SchemeAttributes = 30 // set foreground color to black               // 30	черный
	CLR_FG_RED     SchemeAttributes = 31 // set foreground color to red                 // 31	красный
	CLR_FG_GREEN   SchemeAttributes = 32 // set foreground color to green               // 32	зеленый
	CLR_FG_YELLOW  SchemeAttributes = 33 // set foreground color to yellow              // 33	желтый
	CLR_FG_BLUE    SchemeAttributes = 34 // set foreground color to blue                // 34	синий
	CLR_FG_MAGENTA SchemeAttributes = 35 // set foreground color to magenta (purple)    // 35	пурпурный
	CLR_FG_CYAN    SchemeAttributes = 36 // set foreground color to cyan                // 36	голубой
	CLR_FG_WHITE   SchemeAttributes = 37 // set foreground color to white               // 37	белый
	CLR_FG_DEFAULT SchemeAttributes = 39 // set foreground color to default (white)     // 39 по-умолчанию (белый)

	// BACKGROUND COLOR                                                        // цвет фона
	CLR_BG_BLACK   SchemeAttributes = 40 // set background color to black               // 40	черный
	CLR_BG_RED     SchemeAttributes = 41 // set background color to red                 // 41	красный
	CLR_BG_GREEN   SchemeAttributes = 42 // set background color to green               // 42	зеленый
	CLR_BG_YELLOW  SchemeAttributes = 43 // set background color to yellow              // 43	желтый
	CLR_BG_BLUE    SchemeAttributes = 44 // set background color to blue                // 44	синий
	CLR_BG_MAGENTA SchemeAttributes = 45 // set background color to magenta (purple)    // 45	пурпурный
	CLR_BG_CYAN    SchemeAttributes = 46 // set background color to cyan                // 46	голубой
	CLR_BG_WHITE   SchemeAttributes = 47 // set background color to white               // 47	белый
	CLR_BG_DEFAULT SchemeAttributes = 49 // set background color to default (black)     // 49 по-умолчанию (черный)
)

func copyDefaultCSM(csm *CSM) {
	for key, value := range DefaultCSM {
		(*csm)[key] = value
	}
}

func CSMthemePicker(ThemeName string) (cs CSM) {
	cs = make(CSM)
	copyDefaultCSM(&cs)
	switch ThemeName {
	case "arseny":
		cs[CS_DATETIME] = ColorSheme{CreateCS(CLR_FG_DEFAULT)}
		cs[CS_INFO] = ColorSheme{CreateCS(CLR_FG_GREEN, CLR_BOLD)}
		cs[CS_DEBUG] = ColorSheme{CreateCS(CLR_FG_YELLOW)}
		cs[CS_WARNING] = ColorSheme{CreateCS(CLR_FG_YELLOW, CLR_BOLD)}
		cs[CS_ERROR] = ColorSheme{CreateCS(CLR_FG_RED, CLR_BOLD)}
		cs[CS_FATAL_ERROR] = ColorSheme{CreateCS(CLR_FG_RED, CLR_BOLD, CLR_STRIKE)}
		cs[CS_TEST] = ColorSheme{CreateCS(CLR_FG_BLUE)}
		cs[CS_PRINT] = ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD)}
		cs[CS_DB] = ColorSheme{CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_CYAN, CLR_BOLD)}
		cs[CS_DB_OK] = ColorSheme{CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(CLR_FG_GREEN, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_CYAN, CLR_BOLD)}
		cs[CS_DB_FAIL] = ColorSheme{CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(CLR_FG_RED, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_CYAN, CLR_BOLD)}
		cs[CS_REDIS] = ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD), CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_MAGENTA, CLR_BOLD)}
		cs[CS_REDIS_OK] = ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD), CreateCS(CLR_FG_GREEN, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_MAGENTA, CLR_BOLD)}
		cs[CS_REDIS_FAIL] = ColorSheme{CreateCS(CLR_FG_MAGENTA, CLR_BOLD), CreateCS(CLR_FG_RED, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_MAGENTA, CLR_BOLD)}
		cs[CS_MEMCHD] = ColorSheme{CreateCS(CLR_FG_GREEN, CLR_BOLD), CreateCS(CLR_FG_BLUE, CLR_BOLD), CreateCS(NEW_LINE, CLR_BG_GREEN, CLR_BOLD)}
		return
	default:
		return
	}
}
