package main

import (
	"github.com/gonutz/di8"
	"syscall"
	"time"
	"unsafe"
)

// controlKeys is a list of keys that has to be pressed at the same time for the
// program to check the arrow keys. If all control keys are down and at the same
// time for example the Left and Down arrow keys are pressed, the active window
// will be moved to the bottom left quarter of the screen.
// Note that this does not deactivate the Windows key shortcuts, so if you are
// on a Windows system that uses Windows Key+Left to move the active window to
// the left half of the screen, you should not simply use the windows key for
// these control keys. Use a key combination that will not intercept your usual
// workflow.
var controlKeys = []uint32{di8.K_LCONTROL, di8.K_LWIN}

const (
	SWP_NOSIZE         = 0x0001
	SWP_NOMOVE         = 0x0002
	SWP_NOZORDER       = 0x0004
	SWP_NOREDRAW       = 0x0008
	SWP_NOACTIVATE     = 0x0010
	SWP_DRAWFRAME      = 0x0020
	SWP_FRAMECHANGED   = 0x0020
	SWP_SHOWWINDOW     = 0x0040
	SWP_HIDEWINDOW     = 0x0080
	SWP_NOCOPYBITS     = 0x0100
	SWP_NOOWNERZORDER  = 0x0200
	SWP_NOREPOSITION   = 0x0200
	SWP_NOSENDCHANGING = 0x0400
	SWP_DEFERERASE     = 0x2000
	SWP_ASYNCWINDOWPOS = 0x4000

	SW_HIDE            = 0
	SW_SHOWNORMAL      = 1
	SW_SHOWMINIMIZED   = 2
	SW_MAXIMIZE        = 3
	SW_SHOWMAXIMIZED   = 3
	SW_SHOWNOACTIVATE  = 4
	SW_SHOW            = 5
	SW_MINIMIZE        = 6
	SW_SHOWMINNOACTIVE = 7
	SW_SHOWNA          = 8
	SW_RESTORE         = 9
	SW_SHOWDEFAULT     = 10
	SW_FORCEMINIMIZE   = 11

	MONITOR_DEFAULTTONULL    = 0x00000000
	MONITOR_DEFAULTTOPRIMARY = 0x00000001
	MONITOR_DEFAULTTONEAREST = 0x00000002
)

// load Windows functions
var (
	user   = syscall.NewLazyDLL("user32.dll")
	kernel = syscall.NewLazyDLL("kernel32.dll")

	getModuleHandle = kernel.NewProc("GetModuleHandleW")

	getForegroundWindow = user.NewProc("GetForegroundWindow")
	setWindowPos        = user.NewProc("SetWindowPos")
	showWindow          = user.NewProc("ShowWindow")
	monitorFromWindow   = user.NewProc("MonitorFromWindow")
	getMonitorInfo      = user.NewProc("GetMonitorInfoW")
)

func main() {
	// initialize DirectInput
	check(di8.Init())
	defer di8.Close()

	moduleHandle, _, _ := getModuleHandle.Call(0)
	dinput, err := di8.Create(unsafe.Pointer(moduleHandle))
	check(err)
	defer dinput.Release()

	keyboard, err := dinput.CreateDevice(di8.GUID_SysKeyboard)
	check(err)
	defer keyboard.Release()

	check(keyboard.SetCooperativeLevel(
		nil,
		// receive data even when the application is in the background
		di8.SCL_NONEXCLUSIVE|di8.SCL_BACKGROUND),
	)
	check(keyboard.SetPredefinedDataFormat(di8.DataFormatKeyboard))
	check(keyboard.SetPredefinedProperty(
		di8.PROP_BUFFERSIZE,
		di8.NewPropDword(0, di8.PH_DEVICE, 32),
	))
	check(keyboard.Acquire())
	defer keyboard.Unacquire()

	var (
		controlKeyDown = make([]bool, len(controlKeys))
		escapeKeyDown  bool
		leftKeyDown    bool
		rightKeyDown   bool
		upKeyDown      bool
		downKeyDown    bool
	)

	for {
		time.Sleep(250 * time.Millisecond)
		data, err := keyboard.GetDeviceData(32)
		check(err)
		for i := range data {
			down := data[i].Data != 0
			for j, key := range controlKeys {
				if data[i].Ofs == key {
					controlKeyDown[j] = down
				}
			}
			switch data[i].Ofs {
			case di8.K_LEFT:
				leftKeyDown = down
			case di8.K_RIGHT:
				rightKeyDown = down
			case di8.K_UP:
				upKeyDown = down
			case di8.K_DOWN:
				downKeyDown = down
			case di8.K_ESCAPE:
				escapeKeyDown = down
			}

			// check if action is to be taken
			allControlKeysDown := true
			for _, down := range controlKeyDown {
				if !down {
					allControlKeysDown = false
					break
				}
			}
			if allControlKeysDown {
				if escapeKeyDown {
					return
				}
				if leftKeyDown && upKeyDown {
					reposition(topLeft)
				} else if leftKeyDown && downKeyDown {
					reposition(bottomLeft)
				} else if rightKeyDown && upKeyDown {
					reposition(topRight)
				} else if rightKeyDown && downKeyDown {
					reposition(bottomRight)
				}
			}
		}
	}
}

func reposition(pos position) {
	window, _, _ := getForegroundWindow.Call()
	monitor, _, _ := monitorFromWindow.Call(window, MONITOR_DEFAULTTONEAREST)
	info := MONITORINFO{CbSize: 40}
	ret, _, _ := getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		panic("getMonitorInfo failed")
	}
	showWindow.Call(window, SW_RESTORE)
	r := info.RcWork
	w, h := r.Width()/2, r.Height()/2
	var x, y int32
	switch pos {
	case topLeft:
		x, y = r.Left, r.Top
	case topRight:
		x, y = r.Left+w, r.Top
	case bottomLeft:
		x, y = r.Left, r.Top+h
	case bottomRight:
		x, y = r.Left+w, r.Top+h
	}
	setWindowPos.Call(window, 0,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		SWP_ASYNCWINDOWPOS|SWP_NOACTIVATE|SWP_NOOWNERZORDER|
			SWP_NOZORDER|SWP_SHOWWINDOW)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

func (r RECT) Width() int32 {
	return r.Right - r.Left
}

func (r RECT) Height() int32 {
	return r.Bottom - r.Top
}

type position int

const (
	topLeft position = iota
	topRight
	bottomLeft
	bottomRight
)
