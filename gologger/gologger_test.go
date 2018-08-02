package gologger

import (
	"github.com/imperiuse/golang_lib/colormap"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestPrint(t *testing.T) {
	f, _ := os.Create("log")
	Log := NewLogger(os.Stdout, OFF_ALL, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	defer Log.Close()

	Log.Info("YOU MUST NOT SEE THIS TEXT!!!") // Output: 123
	Log.Debug("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Error("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Test("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Print("YOU MUST NOT SEE THIS TEXT!!!")

	time.Sleep(time.Millisecond * 10)

	newDestinations := GetDefaultDestinations()
	newDestinations[INFO] = []io.Writer{NoColor: ioutil.Discard, Color: io.MultiWriter(f, os.Stdout)}
	newDestinations[DEBUG] = []io.Writer{NoColor: os.Stdout, Color: os.Stdout}
	newDestinations[ERROR] = []io.Writer{NoColor: f, Color: os.Stdout}

	Log.SetNewDestinations(newDestinations)

	Log.Info("INFO MSG", "I", "N", "F", "0", 0)
	Log.Debug("DEBUG MSG", "DE", "BU", "G", new(int))
	Log.Error("ERROR MSG", "!!!!", struct{}{})

	time.Sleep(time.Millisecond * 10)

	Log.SetNewDestinations(GetDefaultDestinations())
	Log.Info("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Debug("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Error("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Test("YOU MUST NOT SEE THIS TEXT!!!")
	Log.Print("YOU MUST NOT SEE THIS TEXT!!!")
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, ON_COLOR, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's Color Info msg!", "Info")
	Log.Debug("It's Color Debug msg!", "Debug")
	Log.Error("It's Color Error msg!", "Error")
	Log.Print("It's Color Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, ON_NO_COLOR, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's  No Color Info msg!", "Info")
	Log.Debug("It's No Color Debug msg!", "Debug")
	Log.Error("It's No Color Error msg!", "Error")
	Log.Print("It's No Color Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, ON_ALL, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's  Both NO Color and COLOR :) Info msg!", "Info")
	Log.Debug("It's Both No Color and COLOR :) Debug msg!", "Debug")
	Log.Error("It's Both No Color and COLOR :) Error msg!", "Error")
	Log.Print("It's Both No Color and COLOR :) Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log.DisableDestinationLvl(INFO)
	Log.Info("YOU MUST NOT SEE THIS TEXT!!!")

	Log.DisableDestinationLvlColor(PRINT, Color)
	Log.Print("ONLY NO COLOR!")

	Log.DisableDestinationLvlColor(TEST, NoColor)
	Log.Test("ONLY COLOR!")

	Log.EnableDestinationLvlColor(INFO, Color)
	Log.Info("Enable only COLOR info!")

	Log.SetAndEnableDestinationLvlColor(ERROR, Color, ioutil.Discard)
	Log.Error("MUST NOT SEE THISE TEXT ERROR DISCARD!")

	Log.SetAndEnableDestinationLvl(ERROR, []io.Writer{ioutil.Discard, ioutil.Discard})
	Log.Error("COLOR ERROR!")

	time.Sleep(time.Millisecond * 10)
}

func BenchmarkNewBasicLogger(b *testing.B) {
	//f, _ := os.Create("log")
	Log := NewLogger(os.Stdout, ON_NO_COLOR, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	for i := 0; i < b.N; i++ {
		Log.Info("My name: ", "!!", "Tick: ", i, "   F1", "F2", "F3", "F4")
	}
}
