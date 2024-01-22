package main

import (
	"fmt"
	"os"
	"regexp"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procEnumWindows          = user32.NewProc("EnumWindows")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procIsWindowVisible      = user32.NewProc("IsWindowVisible")
)

type Window struct {
	Handle syscall.Handle
	Title  string
}

// 获取所有可见窗口的句柄和标题
func getVisibleWindowNamesAndHandles() []Window {
	var windows []Window
	cb := syscall.NewCallback(func(hwnd syscall.Handle, lparam uintptr) uintptr {
		isVisible, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
		if isVisible == 0 {
			return 1 // Skip non-visible windows
		}

		length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
		if length == 0 {
			return 1 // Skip windows with no title
		}

		buffer := make([]uint16, length+1)
		procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buffer[0])), uintptr(length+1))
		title := syscall.UTF16ToString(buffer)
		if title != "" {
			windows = append(windows, Window{Handle: hwnd, Title: title})
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return windows
}

// 根据标题正则表达式查找窗口
func FindWindowByTitle(titleReg string) (Window, error) {
	windows := getVisibleWindowNamesAndHandles()
	for _, window := range windows {
		matched, _ := regexp.MatchString(titleReg, window.Title)
		if matched {
			return window, nil
		}
	}
	return Window{}, fmt.Errorf("window %s not found between %v", titleReg, windows)
}

var (
	getClientRect          = user32.NewProc("GetClientRect")
	clientToScreen         = user32.NewProc("ClientToScreen")
	shcore                 = syscall.NewLazyDLL("shcore.dll")
	setProcessDpiAwareness = shcore.NewProc("SetProcessDpiAwareness")
)

const (
	GWLP_HINSTANCE                = -6
	GWLP_HWNDPARENT               = -8
	GWL_STYLE                     = -16
	PROCESS_PER_MONITOR_DPI_AWARE = 2
)

type Recti32 struct {
	x1, y1, x2, y2 int32
}

type POINT struct {
	X, Y int32
}

// 获取窗口客户区域的矩形
func GetClientAreaRect(hwnd syscall.Handle) Recti32 {
	var rect Recti32
	setProcessDpiAwareness.Call(uintptr(PROCESS_PER_MONITOR_DPI_AWARE))

	getClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rect)))

	var point POINT
	point.X = rect.x1
	point.Y = rect.y1
	clientToScreen.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&point)))

	rect.x1 = point.X
	rect.y1 = point.Y
	rect.x2 += point.X
	rect.y2 += point.Y

	return rect
}

var (
	procSetCursorPos     = user32.NewProc("SetCursorPos")
	procMouseEvent       = user32.NewProc("mouse_event")
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004
	procGetCursorPos     = user32.NewProc("GetCursorPos")
)

func setCursorPos(x, y int) {
	procSetCursorPos.Call(uintptr(x), uintptr(y))
}

func mouseClick(delayms int) {
	procMouseEvent.Call(uintptr(MOUSEEVENTF_LEFTDOWN), 0, 0, 0, 0)
	time.Sleep(time.Duration(delayms) * time.Millisecond)
	procMouseEvent.Call(uintptr(MOUSEEVENTF_LEFTUP), 0, 0, 0, 0)
	time.Sleep(time.Duration(delayms) * time.Millisecond)
}

func getMousePos() (int, int) {
	var point POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	return int(point.X), int(point.Y)
}

func ClickAndBack(x int, y int, delayms int) {
	oldX, oldY := getMousePos()
	setCursorPos(x, y)
	mouseClick(delayms)
	setCursorPos(oldX, oldY)
}

func isAdministrator() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}
	return true
}

func runAsAdministrator() {
	// cmd := exec.Command("cmd", "/C", "start", "AshEchoesAutoPause.exe")
	// cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	// cmd.Run()
}
