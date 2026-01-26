package main

import (
	"fmt"
	"log"
	"runtime"

	"voxel-game/internal/camera"
	"voxel-game/internal/input"
	"voxel-game/internal/player"
	"voxel-game/internal/render"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	windowWidth  = 1280
	windowHeight = 720
	windowTitle  = "Simple Voxel Game"
)

func init() {
	// GLFW requires this to run on main thread
	runtime.LockOSThread()
}

func main() {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// Configure GLFW
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}
	window.MakeContextCurrent()

	// Control VSync
	glfw.SwapInterval(0) // 1 = VSync on, 0 = VSync off

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln("failed to initialize OpenGL:", err)
	}

	// Configure OpenGL
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.ClearColor(0.53, 0.81, 0.92, 1.0) // Sky blue

	// Print OpenGL version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version:", version)

	// Initialize renderer
	renderer, err := render.NewRenderer()
	if err != nil {
		log.Fatalln("failed to create renderer:", err)
	}

	wireframeMode := false

	// Initialize camera
	cam := camera.NewCamera(windowWidth, windowHeight)

	// Initialize world
	gameWorld := world.NewWorld()

	// Initialize player
	p := player.NewPlayer(cam, gameWorld)

	// Initialize input manager
	inputMgr := input.NewInputManager(window, cam, p, &wireframeMode)

	// Capture cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Track delta time
	lastTime := glfw.GetTime()

	// FPS tracking
	frameCount := 0
	fpsTime := glfw.GetTime()
	currentFPS := 0.0

	// Chunk update throttling
	lastChunkUpdate := glfw.GetTime()
	chunkUpdateInterval := 1.0 // Update chunks every 0.5 seconds

	// Game loop
	for !window.ShouldClose() {
		// Calculate delta time
		currentTime := glfw.GetTime()
		deltaTime := float32(currentTime - lastTime)
		lastTime = currentTime

		// FPS calculation
		frameCount++
		if currentTime-fpsTime >= 1.0 {
			currentFPS = float64(frameCount) / (currentTime - fpsTime)
			window.SetTitle(fmt.Sprintf("%s - FPS: %.1f", windowTitle, currentFPS))
			frameCount = 0
			fpsTime = currentTime
		}

		// Handle input
		inputMgr.Update(deltaTime)

		// Update player
		p.Update(deltaTime)

		// Update world chunks based on player position
		if currentTime-lastChunkUpdate >= chunkUpdateInterval {
			gameWorld.UpdateChunks(cam.Position[0], cam.Position[2])
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render world
		renderer.RenderWorld(gameWorld, cam)

		// Swap buffers and poll events
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
