package main

import (
	"strconv"
	"syscall/js"
	"time"

	"github.com/55utah/fc-simulator/nes"
)

var console *nes.Console

var timestamp float64

// dom canvas ctx对象
var ctx js.Value
var canvas js.Value
var document js.Value
var window js.Value
var navigator js.Value

var ratio = 1
var width = 256
var height = 240

var ctrl1 [8]bool
var ctrl2 [8]bool

// 写死一个最大值
const FrameSecond float64 = 0.0167

const (
	keyA int = iota
	keyB
	keySelect
	keyStart
	keyUp
	keyDown
	keyLeft
	keyRight
	keyInvaild int = -1
)

func floatSecond() float64 {
	return float64(time.Now().Nanosecond()) * float64(1e-9)
}

// 这个缩放会导致帧率下降，所以先去掉
// func resize(source *image.RGBA, w int, h int, ratio int) *image.RGBA {
// 	if ratio == 1 {
// 		return source
// 	}

// 	tw := w * ratio
// 	th := h * ratio

// 	var target *image.RGBA = image.NewRGBA(image.Rect(0, 0, tw, th))

// 	for y := 0; y < th; y++ {
// 		for x := 0; x < tw; x++ {
// 			sx := x / ratio
// 			sy := y / ratio
// 			target.SetRGBA(x, y, source.RGBAAt(sx, sy))
// 		}
// 	}

// 	return target
// }

// 将js更新canvas改为go修改dom，执行更快，cpu占用更小！！
func render(value []byte, width int, height int) {
	imageData := ctx.Call("getImageData", 0, 0, width, height)
	buf := js.Global().Get("Uint8ClampedArray").New(width * height * 4)

	dst := js.Global().Get("Uint8Array").New(len(value))
	js.CopyBytesToJS(dst, value)
	buf.Call("set", dst)
	imageData.Get("data").Call("set", buf)
	ctx.Call("putImageData", imageData, 0, 0)
}

func onFrame() {
	gamepads := navigator.Call("getGamepads")
	current := floatSecond()
	cost := current - timestamp
	if cost > FrameSecond {
		cost = FrameSecond
	} else if cost < 0 {
		cost = 0
	}

	gamepad1 := gamepads.Index(0)
	if gamepad1.Type() != js.TypeUndefined && gamepad1.Type() != js.TypeNull {
		buttons := gamepad1.Get("buttons")
		axes := gamepad1.Get("axes")
		if buttons.Length() == 11 && axes.Length() == 8 {
			// for Microsoft Xbox One X pad
			ctrl1[keyA] = buttons.Index(0).Get("pressed").Bool()
			ctrl1[keyB] = buttons.Index(1).Get("pressed").Bool()
			ctrl1[keySelect] = buttons.Index(6).Get("pressed").Bool()
			ctrl1[keyStart] = buttons.Index(7).Get("pressed").Bool()
			axesX := axes.Index(6).Int()
			axesY := axes.Index(7).Int()
			ctrl1[keyUp] = (axesY == -1)
			ctrl1[keyDown] = (axesY == 1)
			ctrl1[keyLeft] = (axesX == -1)
			ctrl1[keyRight] = (axesX == 1)
			console.SetButton1(ctrl1)
		}
	}

	gamepad2 := gamepads.Index(1)
	if gamepad2.Type() != js.TypeUndefined && gamepad2.Type() != js.TypeNull {
		buttons := gamepad2.Get("buttons")
		axes := gamepad2.Get("axes")
		if buttons.Length() == 11 && axes.Length() == 8 {
			// for Microsoft Xbox One X pad
			ctrl2[keyA] = buttons.Index(0).Get("pressed").Bool()
			ctrl2[keyB] = buttons.Index(1).Get("pressed").Bool()
			ctrl2[keySelect] = buttons.Index(6).Get("pressed").Bool()
			ctrl2[keyStart] = buttons.Index(7).Get("pressed").Bool()
			axesX := axes.Index(6).Int()
			axesY := axes.Index(7).Int()
			ctrl2[keyUp] = (axesY == -1)
			ctrl2[keyDown] = (axesY == 1)
			ctrl2[keyLeft] = (axesX == -1)
			ctrl2[keyRight] = (axesX == 1)
			console.SetButton2(ctrl2)
		}
	}

	timestamp = current
	console.StepSeconds(cost)
	buffer := console.Buffer()
	// 放弃数据缩放，改为canvas样式缩放
	render(buffer.Pix, width, height)
}

func global() js.Value {
	return js.Global().Get("customConsole")
}

// const BufferSize = 8192
const BufferSize = 2048

var audioFloatArray = make([]float32, BufferSize)
var audioIndex = 0

func outputAudio(v float32) {
	if audioIndex < BufferSize-1 {
		audioFloatArray[audioIndex] = v
		audioIndex++
	} else {
		jsFloatArray := js.Global().Get("Float32Array").New(BufferSize)

		for index, val := range audioFloatArray {
			jsFloatArray.SetIndex(index, val)
		}
		global().Set("audioFloa32Array", jsFloatArray)

		audioIndex = 0
		audioFloatArray = make([]float32, BufferSize)
	}
}

func newConsole(file js.Value, sampleRate int) {
	game := make([]byte, file.Get("length").Int())
	js.CopyBytesToGo(game, file)
	console, _ = nes.NewConsole(game)
	console.SetAudioSampleRate(float64(sampleRate))
	console.SetAudioOutputWork(outputAudio)

	window = js.Global().Get("window")
	document = js.Global().Get("document")
	navigator = js.Global().Get("navigator")
	canvas = document.Call("querySelector", "canvas")
	ctx = canvas.Call("getContext", "2d")

	handleKeyEvent()
	handleGamepadEvent()
}

func handleKeyEvent() {
	onkeydownCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		handleKey(e.Get("code").String(), true)
		return nil
	})
	document.Set("onkeydown", onkeydownCallback)

	onkeyupCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		handleKey(e.Get("code").String(), false)
		return nil
	})
	document.Set("onkeyup", onkeyupCallback)
}

func handleGamepadEvent() {
	onGamepadConnCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		println("Gamepad connected: " + e.Get("gamepad").Get("id").String())
		return nil
	})
	window.Call("addEventListener", "gamepadconnected", onGamepadConnCallback)

	onGamepadDisconnCallback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		e := args[0]
		println("Gamepad disconnect: " + e.Get("gamepad").Get("id").String())
		return nil
	})
	window.Call("addEventListener", "gamepaddisconnected", onGamepadDisconnCallback)
}

func handleKey(code string, down bool) {

	if console == nil {
		return
	}

	if down {
		keyParseSys(code, func() {
			canvas.Get("style").Set("width", strconv.Itoa(width*ratio)+"px")
			canvas.Get("style").Set("height", strconv.Itoa(height*ratio)+"px")
		})
	}

	index1 := keyPress1(code)
	index2 := keyPress2(code)

	if index1 >= 0 {
		ctrl1[index1] = down
		console.SetButton1(ctrl1)
	}
	if index2 >= 0 {
		ctrl2[index2] = down
		console.SetButton2(ctrl2)
	}
}

func main() {
	println("Hello, fc!")

	// newConsole
	global().Set("newConsole", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		newConsole(args[0], args[1].Int())
		return nil
	}))

	// onFrame
	global().Set("frame", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		onFrame()
		return nil
	}))

	channel := make(chan int, 0)
	<-channel
	println("finished!")
}

func keyParseSys(keyName string, resizeWindow func()) {
	switch keyName {
	// 重置游戏
	case "KeyQ":
		console.Reset()
	// 缩小屏幕
	case "Minus":
		if ratio > 1 {
			ratio--
			resizeWindow()
		}
	// 放大屏幕
	case "Equal":
		if ratio < 4 {
			ratio++
			resizeWindow()
		}
	}
}

func keyPress1(code string) int {
	switch code {
	// A
	case "KeyF":
		return keyA
	// B
	case "KeyG":
		return keyB
	// Select
	case "KeyR":
		return keySelect
	// Start
	case "KeyT":
		return keyStart
	// 上
	case "KeyW":
		return keyUp
	// 下
	case "KeyS":
		return keyDown
	// 左
	case "KeyA":
		return keyLeft
	// 右
	case "KeyD":
		return keyRight
	}
	return keyInvaild
}

func keyPress2(code string) int {
	switch code {
	// A
	case "KeyJ":
		return keyA
	// B
	case "KeyK":
		return keyB
	// Select
	case "KeyU":
		return keySelect
	// Start
	case "KeyI":
		return keyStart
	// 上
	case "ArrowUp":
		return keyUp
	// 下
	case "ArrowDown":
		return keyDown
	// 左
	case "ArrowLeft":
		return keyLeft
	// 右
	case "ArrowRight":
		return keyRight
	}
	return keyInvaild
}
