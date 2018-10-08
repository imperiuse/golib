package colors	
 import (	
	"fmt"	
	"github.com/imperiuse/golib/colormap"
)	
 var (	
	RED     string	
	GREEN   string	
	BLUE    string	
	MAGENTA string	
	YELLOW  string	
	BLACK   string	
	NBLACK  string	
	RESET   string	
)	
 func init() {	
	RED = colormap.CreateCS(colormap.CLR_FG_RED, colormap.CLR_BOLD)	
	GREEN = colormap.CreateCS(colormap.CLR_FG_GREEN, colormap.CLR_BOLD)	
	BLUE = colormap.CreateCS(colormap.CLR_FG_BLUE, colormap.CLR_BOLD)	
	MAGENTA = colormap.CreateCS(colormap.CLR_FG_MAGENTA, colormap.CLR_BOLD)	
	YELLOW = colormap.CreateCS(colormap.CLR_FG_YELLOW, colormap.CLR_BOLD)	
	NBLACK = colormap.CreateCS(colormap.CLR_BG_BLACK, colormap.CLR_BOLD)	
	BLACK = colormap.CreateCS(colormap.CLR_FG_BLACK, colormap.CLR_BOLD)	
	RESET = colormap.CreateCS(colormap.CLR_RESET)	
}	
 // Функция для логирования ошибки или статуса что ее нет.	
func CheckErrorFunc(err error, f string) {	
	if err != nil {	
		fmt.Println("[CheckErrFunc]", RED, "Error while ", f,  err, "\n", RESET)	
	} else {	
		fmt.Println("[CheckErrFunc]", GREEN, "Successful ", f, "\n", RESET)	
	}	
}	
 func ColorizedString(color string, value string) string {	
	return fmt.Sprint(color, value, RESET)	
}	
 func ColorizedFloat64(color string, value float64) string {	
	return fmt.Sprintf("%v%.8f%v", color, value, RESET)	
}	
 func Colorized(color string, v interface{}) string {	
	return fmt.Sprintf("%v%v%v", color, v, RESET)	
}	
 func ChooseColorBool(trueColor, falseColor string, v bool) string {	
	if v {	
		return fmt.Sprint(trueColor, v, RESET)	
	} else {	
		return fmt.Sprint(falseColor, v, RESET)	
	}	
}	
 func ChooseColorMoreThanValueInt(moreColor, lowerColor string, moreValue, v int) string {	
	if v > moreValue {	
		return fmt.Sprint(moreColor, v, RESET)	
	} else {	
		return fmt.Sprint(lowerColor, v, RESET)	
	}	
}	
 func ChooseColorMoreThanValueFloat(moreColor, lowerColor string, moreValue, v float64) string {	
	if v > moreValue {	
		return fmt.Sprintf("%v%.8f%v", moreColor, v, RESET)	
	} else {	
		return fmt.Sprintf("%v%.8f%v", lowerColor, v, RESET)	
	}	
}	
 func ChooseColorNonEqValueFloat(moreColor, lowerColor string, moreValue, v float64) string {	
	if v != moreValue {	
		return fmt.Sprintf("%v%.8f%v", moreColor, v, RESET)	
	} else {	
		return fmt.Sprintf("%v%.8f%v", lowerColor, v, RESET)	
	}	
}
