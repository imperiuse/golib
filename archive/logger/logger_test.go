package gologger

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/imperiuse/golib/archive/colormap"
)

func TestPrint(t *testing.T) {
	f, _ := os.Create("log")
	Log := NewLogger(os.Stdout, OffAll, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	defer Log.Close()

	Log.Info("This is printed text")  // Output: This is printed text
	Log.Debug("This is printed text") // Output: This is printed text
	Log.Error("This is printed text") // Output: This is printed text
	Log.Test("This is printed text")  // Output: This is printed text
	Log.Print("This is printed text") // Output: This is printed text

	time.Sleep(time.Millisecond * 10)

	newDestinations := GetDefaultDestinations()
	newDestinations[Info] = []io.Writer{NoColor: ioutil.Discard, Color: io.MultiWriter(f, os.Stdout)}
	newDestinations[Debug] = []io.Writer{NoColor: os.Stdout, Color: os.Stdout}
	newDestinations[Error] = []io.Writer{NoColor: f, Color: os.Stdout}

	Log.SetNewDestinations(newDestinations)

	Log.Info("Info MSG", "I", "N", "F", "0", 0)
	Log.Debug("Debug MSG", "DE", "BU", "G", new(int))
	Log.Error("Error MSG", "!!!!", struct{}{})

	time.Sleep(time.Millisecond * 10)

	Log.SetNewDestinations(GetDefaultDestinations())
	Log.Info("This is printed text")  // Output: This is printed text
	Log.Debug("This is printed text") // Output: This is printed text
	Log.Error("This is printed text") // Output: This is printed text
	Log.Test("This is printed text")  // Output: This is printed text
	Log.Print("This is printed text") // Output: This is printed text
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, OnColor, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's Color Info msg!", "Info")
	Log.Debug("It's Color Debug msg!", "Debug")
	Log.Error("It's Color Error msg!", "Error")
	Log.Print("It's Color Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, OnNoColor, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's  No Color Info msg!", "Info")
	Log.Debug("It's No Color Debug msg!", "Debug")
	Log.Error("It's No Color Error msg!", "Error")
	Log.Print("It's No Color Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log = NewLogger(os.Stdout, OnAll, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	Log.Info("It's  Both NO Color and COLOR :) Info msg!", "Info")
	Log.Debug("It's Both No Color and COLOR :) Debug msg!", "Debug")
	Log.Error("It's Both No Color and COLOR :) Error msg!", "Error")
	Log.Print("It's Both No Color and COLOR :) Print msg!", "Print")
	Log.P()
	time.Sleep(time.Millisecond * 10)

	Log.DisableDestinationLvl(Info)
	Log.Info("YOU MUST NOT SEE THIS TEXT!!!")

	Log.DisableDestinationLvlColor(Print, Color)
	Log.Print("ONLY NO COLOR!")

	Log.DisableDestinationLvlColor(Test, NoColor)
	Log.Test("ONLY COLOR!")

	Log.EnableDestinationLvlColor(Info, Color)
	Log.Info("Enable only COLOR info!")

	Log.SetAndEnableDestinationLvlColor(Error, Color, ioutil.Discard)
	Log.Error("MUST NOT SEE THISE TEXT Error DISCARD!")

	Log.SetAndEnableDestinationLvl(Error, []io.Writer{ioutil.Discard, ioutil.Discard})
	Log.Error("COLOR Error!")

	time.Sleep(time.Millisecond * 10)
}

func BenchmarkNewBasicLogger(b *testing.B) {
	//f, _ := os.Create("log")
	Log := NewLogger(os.Stdout, OnNoColor, 100, 0, Ldate|Ltime|Lshortfile, "\t",
		colormap.CSMthemePicker("arseny"))

	for i := 0; i < b.N; i++ {
		Log.Info("My name: ", "!!", "Tick: ", i, "   F1", "F2", "F3", "F4")
	}
}
