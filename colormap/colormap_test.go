package colormap

import "testing"

// nolint
func TestCreateCS(t *testing.T) {
	if CreateCS(ClrFgBlack) != "\x1b[30m" {
		t.Fail()
	}
	if CreateCS(ClrFgRed) != "\x1b[31m" {
		t.Fail()
	}
	if CreateCS(ClrFgGreen) != "\x1b[32m" {
		t.Fail()
	}
	if CreateCS(ClrFgYellow) != "\x1b[33m" {
		t.Fail()
	}
	if CreateCS(ClrFgBlue) != "\x1b[34m" {
		t.Fail()
	}
	if CreateCS(ClrFgMagenta) != "\x1b[35m" {
		t.Fail()
	}
	if CreateCS(ClrFgCyan) != "\x1b[36m" {
		t.Fail()
	}
	if CreateCS(ClrFgWhite) != "\x1b[37m" {
		t.Fail()
	}
	if CreateCS(ClrFgDefault) != "\x1b[39m" {
		t.Fail()
	}

	if CreateCS(ClrBgBlack) != "\x1b[40m" {
		t.Fail()
	}
	if CreateCS(ClrBgRed) != "\x1b[41m" {
		t.Fail()
	}
	if CreateCS(ClrBgGreen) != "\x1b[42m" {
		t.Fail()
	}
	if CreateCS(ClrBgYellow) != "\x1b[43m" {
		t.Fail()
	}
	if CreateCS(ClrBgBlue) != "\x1b[44m" {
		t.Fail()
	}
	if CreateCS(ClrBgMagenta) != "\x1b[45m" {
		t.Fail()
	}
	if CreateCS(ClrBgCyan) != "\x1b[46m" {
		t.Fail()
	}
	if CreateCS(ClrBgWhite) != "\x1b[47m" {
		t.Fail()
	}
	if CreateCS(ClrBgDefault) != "\x1b[49m" {
		t.Fail()
	}

	if CreateCS(ClrReset) != "\x1b[0m" {
		t.Fail()
	}
	if CreateCS(ClrBold) != "\x1b[1m" {
		t.Fail()
	}
	if CreateCS(ClrItal) != "\x1b[3m" {
		t.Fail()
	}
	if CreateCS(ClrUnder) != "\x1b[4m" {
		t.Fail()
	}
	if CreateCS(ClrFlashing) != "\x1b[5m" {
		t.Fail()
	}
	if CreateCS(ClrInverse) != "\x1b[7m" {
		t.Fail()
	}
	if CreateCS(ClrInvisible) != "\x1b[8m" {
		t.Fail()
	}
	if CreateCS(ClrStrike) != "\x1b[9m" {
		t.Fail()
	}

	if CreateCS(ClrBoldOff) != "\x1b[21m" {
		t.Fail()
	}
	if CreateCS(ClrItalOff) != "\x1b[23m" {
		t.Fail()
	}
	if CreateCS(ClrUnderOff) != "\x1b[24m" {
		t.Fail()
	}
	if CreateCS(ClrFlashingOff) != "\x1b[25m" {
		t.Fail()
	}
	if CreateCS(ClrInversOff) != "\x1b[27m" {
		t.Fail()
	}
	if CreateCS(ClrInvisibleOff) != "\x1b[28m" {
		t.Fail()
	}
	if CreateCS(ClrStrikeOff) != "\x1b[29m" {
		t.Fail()
	}
}
