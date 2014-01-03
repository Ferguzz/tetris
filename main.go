package main

import (
	"errors"
	"fmt"
	"github.com/Ferguzz/glam"
	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"runtime"
	"time"
)

// Window stuff.
const (
	height   = 640.0
	width    = 480.0
	aspect   = width / height
	gridSize = 20
	speed    = 2
)

// Controls.
const (
	down = iota
	left
	right
)

// The falling block.
var shaderProgram gl.Program
var generateNewBlock chan bool = make(chan bool, 1)

func run() {
	// All OpenGL related stuff needs to be called from the same goroutine.  The mainThread function achieves this.
	mainThread(func() {
		window, err := glInit()
		if err != nil {
			panic(err)
		}
		defer glExit()

		vertexShader, err := loadShader(gl.VERTEX_SHADER, block_vertex_shader)
		if err != nil {
			panic(err)
		}
		defer vertexShader.Delete()

		fragmentShader, err := loadShader(gl.FRAGMENT_SHADER, block_fragment_shader)
		if err != nil {
			panic(err)
		}
		defer fragmentShader.Delete()

		shaderProgram = gl.CreateProgram()
		shaderProgram.AttachShader(vertexShader)
		shaderProgram.AttachShader(fragmentShader)
		shaderProgram.BindFragDataLocation(0, "outColor")
		shaderProgram.Link()
		shaderProgram.Use()

		shaderProgram.DetachShader(vertexShader)
		shaderProgram.DetachShader(fragmentShader)
		defer shaderProgram.Delete()

		GenerateShapes()
		defer CleanUpShapes()

		projection := glam.Orthographic(height, aspect, -1, 1)
		shaderProgram.GetUniformLocation("projection").UniformMatrix4fv(false, projection)
		shaderProgram.GetUniformLocation("scale").Uniform1f(gridSize)
		gl.ClearColor(0, 0, 0, 1)

		// Create the first block.
		NewBlock()

		// Set off a thread to cause the active block to fall.
		go func() {
			for _ = range time.Tick(time.Second / speed) {
				blocks[len(blocks)-1].Move(down)
			}
		}()

		for !window.ShouldClose() {
			gl.Clear(gl.COLOR_BUFFER_BIT)

			// Check for messages requesting a new block.
			select {
			case <-generateNewBlock:
				NewBlock()
			default:
			}

			for _, block := range blocks {
				block.Draw()
			}

			window.SwapBuffers()
			glfw.PollEvents()
		}
	})
	close(mainfunc)
}

func glInit() (*glfw.Window, error) {
	glfw.SetErrorCallback(errorCallback)

	if !glfw.Init() {
		return nil, errors.New("Can't initialise GLFW!")
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenglProfile, glfw.OpenglCoreProfile)
	glfw.WindowHint(glfw.OpenglForwardCompatible, gl.TRUE)

	window, err := glfw.CreateWindow(width, height, "Tetris", nil, nil)
	if err != nil {
		return nil, err
	}

	window.SetKeyCallback(keyCallback)
	window.MakeContextCurrent()
	if gl.Init() != 0 {
		return nil, errors.New("Can't initialise OpenGL.")
	}

	return window, nil
}

func glExit() {
	glfw.Terminate()
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.KeyQ:
			window.SetShouldClose(true)
		case glfw.KeyLeft:
			blocks[len(blocks)-1].Move(left)
		case glfw.KeyRight:
			blocks[len(blocks)-1].Move(right)
		case glfw.KeyDown:
			blocks[len(blocks)-1].Move(down)
		case glfw.KeyR:
			blocks[len(blocks)-1].Rotate()
		}
	}
}

func loadShader(shaderType gl.GLenum, source string) (gl.Shader, error) {
	shader := gl.CreateShader(shaderType)

	shader.Source(source)
	shader.Compile()

	if shader.Get(gl.COMPILE_STATUS) == 0 {
		return shader, errors.New(fmt.Sprintf("Shader (%v) did not compile.", source))
	}

	return shader, nil
}

func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.LockOSThread()
}

func Main() {
	for f := range mainfunc {
		f()
	}
}

var mainfunc = make(chan func())

func mainThread(f func()) {
	done := make(chan bool, 1)
	mainfunc <- func() {
		f()
		done <- true
	}
	<-done
}

func main() {
	go run()
	Main()
}
