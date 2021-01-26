package colors

import (
	"fmt"

	"github.com/imperiuse/golib/archive/colormap"
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
	RED = colormap.CreateCS(colormap.ClrFgRed, colormap.ClrBold)
	GREEN = colormap.CreateCS(colormap.ClrFgGreen, colormap.ClrBold)
	BLUE = colormap.CreateCS(colormap.ClrFgBlue, colormap.ClrBold)
	MAGENTA = colormap.CreateCS(colormap.ClrFgMagenta, colormap.ClrBold)
	YELLOW = colormap.CreateCS(colormap.ClrFgYellow, colormap.ClrBold)
	NBLACK = colormap.CreateCS(colormap.ClrBgBlack, colormap.ClrBold)
	BLACK = colormap.CreateCS(colormap.ClrFgBlack, colormap.ClrBold)
	RESET = colormap.CreateCS(colormap.ClrReset)
}

// CheckErrorFunc - logging func of error
func CheckErrorFunc(err error, f string) {
	if err != nil {
		fmt.Println("[CheckErrFunc]", RED, "Error while ", f, err, "\n", RESET)
	} else {
		fmt.Println("[CheckErrFunc]", GREEN, "Successful ", f, "\n", RESET)
	}
}

// ColorizedString - make color string
func ColorizedString(color string, value string) string {
	return fmt.Sprint(color, value, RESET)
}

// ColorizedFloat64 - make color float64
func ColorizedFloat64(color string, value float64) string {
	return fmt.Sprintf("%v%.8f%v", color, value, RESET)
}

// Colorized - make color some string representation of interface
func Colorized(color string, v interface{}) string {
	return fmt.Sprintf("%v%v%v", color, v, RESET)
}

// ChooseColorBool - if colorized
func ChooseColorBool(trueColor, falseColor string, v bool) string {
	if v {
		return fmt.Sprint(trueColor, v, RESET)
	}
	return fmt.Sprint(falseColor, v, RESET)
}

// ChooseColorMoreThanValueInt - if more Colorized
func ChooseColorMoreThanValueInt(moreColor, lowerColor string, moreValue, v int) string {
	if v > moreValue {
		return fmt.Sprint(moreColor, v, RESET)
	}
	return fmt.Sprint(lowerColor, v, RESET)
}

// ChooseColorMoreThanValueFloat -if more Colorized (float)
func ChooseColorMoreThanValueFloat(moreColor, lowerColor string, moreValue, v float64) string {
	if v > moreValue {
		return fmt.Sprintf("%v%.8f%v", moreColor, v, RESET)
	}
	return fmt.Sprintf("%v%.8f%v", lowerColor, v, RESET)
}

// ChooseColorNonEqValueFloat - if not equals colorized
func ChooseColorNonEqValueFloat(moreColor, lowerColor string, moreValue, v float64) string {
	if v != moreValue {
		return fmt.Sprintf("%v%.8f%v", moreColor, v, RESET)
	}
	return fmt.Sprintf("%v%.8f%v", lowerColor, v, RESET)
}
