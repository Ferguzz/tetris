package main

import (
	"github.com/Ferguzz/glam"
	"github.com/go-gl/gl"
	"math"
	"math/rand"
	"time"
)

// Blocks are what are drawn to the screen so they need to remeber where they are and what color they are etc.
type Block struct {
	shape       *Shape
	position    glam.Vec3
	orientation int
	reflect     int
	color       Color
}

// Shapes contain all the information that a block needs to draw itself e.g. vertex arrays, element buffers etc.
type Shape struct {
	vertices      []gl.GLfloat
	elements      []gl.GLushort
	vao           gl.VertexArray
	vbo           gl.Buffer
	elementBuffer gl.Buffer
	numElements   int
}

// This slice holds all the blocks that need to be drawn.
var blocks []Block

const numShapes = 4

var Shapes []Shape = make([]Shape, numShapes)

type Color []float32

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomColor() Color {
	var color Color
	switch rand.Intn(4) {
	case 0:
		// Red.
		color = Color{1, 0, 0}
	case 1:
		// Green.
		color = Color{0, 1, 0}
	case 2:
		// Blue.
		color = Color{0, 0, 1}
	case 3:
		// White.
		color = Color{1, 1, 1}
	}

	return color
}

// Generate all the vaos on startup.  I could do this only when they are first required, but this may delay the block generation.
func GenerateShapes() {
	// Square
	Shapes[0].vertices = []gl.GLfloat{-1, -1, 1, -1, -1, 1, 1, 1}
	Shapes[0].elements = []gl.GLushort{0, 1, 2, 2, 3, 1}

	// ___|
	Shapes[1].vertices = []gl.GLfloat{-2, 0, -2, -1, 2, -1, 2, 0, 2, 1, 1, 1, 1, 0}
	Shapes[1].elements = []gl.GLushort{0, 1, 2, 2, 3, 0, 3, 4, 5, 5, 6, 3}

	// _|_
	Shapes[2].vertices = []gl.GLfloat{-1.5, 0, -0.5, 0, -0.5, 1, 0.5, 1, 0.5, 0, 1.5, 0, 1.5, -1, -1.5, -1}
	Shapes[2].elements = []gl.GLushort{1, 2, 3, 3, 4, 1, 0, 7, 6, 6, 0, 5}

	// Snake
	Shapes[3].vertices = []gl.GLfloat{-1.5, -1, -1.5, 0, -0.5, 0, -0.5, 1, 1.5, 1, 1.5, 0, 0.5, 0, 0.5, -1}
	Shapes[3].elements = []gl.GLushort{0, 1, 6, 6, 7, 0, 2, 3, 4, 4, 5, 2}

	// Now fill out the rest automatically.
	// FIXME why doesn't using _, shape in this loop work ?
	for i := range Shapes {
		Shapes[i].vao = gl.GenVertexArray()
		Shapes[i].vao.Bind()
		Shapes[i].vbo = gl.GenBuffer()
		Shapes[i].vbo.Bind(gl.ARRAY_BUFFER)
		gl.BufferData(gl.ARRAY_BUFFER, len(Shapes[i].vertices)*4, Shapes[i].vertices, gl.STATIC_DRAW)
		Shapes[i].elementBuffer = gl.GenBuffer()
		Shapes[i].elementBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(Shapes[i].elements)*2, Shapes[i].elements, gl.STATIC_DRAW)
		Shapes[i].numElements = len(Shapes[i].elements)

		vertexAttribArray := shaderProgram.GetAttribLocation("position")
		vertexAttribArray.AttribPointer(2, gl.FLOAT, false, 0, uintptr(0))
		vertexAttribArray.EnableArray()
	}
}

func CleanUpShapes() {
	// Fixme does this actually work or is it broken in the same way as above ?
	for _, shape := range Shapes {
		shape.vao.Delete()
		shape.vbo.Delete()
		shape.elementBuffer.Delete()
	}
}

func NewBlock() {
	var block Block
	block.shape = &Shapes[rand.Intn(numShapes)]

	// Initialise a random X position.
	// The blocks must snap to the grid.  Since shapes are defined around their COM, the centres are not what should be
	// snapped.  How do I deal with this ?
	block.position = glam.Vec3{float32((rand.Intn(width/gridSize))*gridSize) - width/2, 200, 0}
	// Pick a random orientation.
	block.orientation = rand.Intn(4)
	block.reflect = rand.Intn(2)

	// Finally, a random color.
	block.color = randomColor()

	blocks = append(blocks, block)
}

func (block *Block) Draw() {
	block.shape.vao.Bind()
	position := glam.Translation(block.position)
	rotation := glam.Rotation(float32(block.orientation)*math.Pi/2.0, glam.Vec3{0, 0, 1})
	var model glam.Mat4
	model.Multiply(&rotation, &position)
	shaderProgram.GetUniformLocation("model").UniformMatrix4fv(false, model)
	shaderProgram.GetUniformLocation("inColor").Uniform3fv(1, block.color)
	// shaderProgram.GetUniformLocation("reflect").Uniform1i(block.reflect)
	gl.DrawElements(gl.TRIANGLES, block.shape.numElements, gl.UNSIGNED_SHORT, uintptr(0))
}

func (block *Block) Move(direction int) {
	// Do collision detection in here.
	switch direction {
	case down:
		if block.position.Y >= -(height/2)+40 {
			block.position.Y -= gridSize
		} else {
			// We've hit the bottom.  Generate a new block.
			// Since block creation involves OpenGL related calls, it needs to be done in the main thread.
			// Since this function can be called from other threads, we need to signal to the OpenGL thread to make a block.
			generateNewBlock <- true
		}
	case left:
		if block.position.X >= -width/2+40 {
			block.position.X -= gridSize
		}
	case right:
		if block.position.X <= width/2-40 {
			block.position.X += gridSize
		}
	}
}

func (block *Block) Rotate() {
	block.orientation += 1
}
