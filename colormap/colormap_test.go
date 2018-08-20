package colormap

import "testing"

func TestCreateCS(t *testing.T) {
	if CreateCS(CLR_FG_BLACK) != "\x1b[30m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_RED) != "\x1b[31m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_GREEN) != "\x1b[32m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_YELLOW) != "\x1b[33m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_BLUE) != "\x1b[34m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_MAGENTA) != "\x1b[35m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_CYAN) != "\x1b[36m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_WHITE) != "\x1b[37m" {
		t.Fail()
	}
	if CreateCS(CLR_FG_DEFAULT) != "\x1b[39m" {
		t.Fail()
	}

	if CreateCS(CLR_BG_BLACK) != "\x1b[40m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_RED) != "\x1b[41m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_GREEN) != "\x1b[42m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_YELLOW) != "\x1b[43m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_BLUE) != "\x1b[44m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_MAGENTA) != "\x1b[45m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_CYAN) != "\x1b[46m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_WHITE) != "\x1b[47m" {
		t.Fail()
	}
	if CreateCS(CLR_BG_DEFAULT) != "\x1b[49m" {
		t.Fail()
	}

	if CreateCS(CLR_RESET) != "\x1b[0m" {
		t.Fail()
	}
	if CreateCS(CLR_BOLD) != "\x1b[1m" {
		t.Fail()
	}
	if CreateCS(CLR_ITAL) != "\x1b[3m" {
		t.Fail()
	}
	if CreateCS(CLR_UNDER) != "\x1b[4m" {
		t.Fail()
	}
	if CreateCS(CLR_FLASHING) != "\x1b[5m" {
		t.Fail()
	}
	if CreateCS(CLR_INVERSE) != "\x1b[7m" {
		t.Fail()
	}
	if CreateCS(CLR_INVISIBLE) != "\x1b[8m" {
		t.Fail()
	}
	if CreateCS(CLR_STRIKE) != "\x1b[9m" {
		t.Fail()
	}

	if CreateCS(CLR_BOLD_OFF) != "\x1b[21m" {
		t.Fail()
	}
	if CreateCS(CLR_ITAL_OFF) != "\x1b[23m" {
		t.Fail()
	}
	if CreateCS(CLR_UNDER_OFF) != "\x1b[24m" {
		t.Fail()
	}
	if CreateCS(CLR_FLASHING_OFF) != "\x1b[25m" {
		t.Fail()
	}
	if CreateCS(CLR_INVERS_OFF) != "\x1b[27m" {
		t.Fail()
	}
	if CreateCS(CLR_INVISIBLE_OFF) != "\x1b[28m" {
		t.Fail()
	}
	if CreateCS(CLR_STRIKE_OFF) != "\x1b[29m" {
		t.Fail()
	}
}
