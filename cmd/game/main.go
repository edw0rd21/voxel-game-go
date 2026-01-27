package main

import (
	"fmt"
	"log"
	"runtime"

	"voxel-game/internal/camera"
	"voxel-game/internal/input"
	"voxel-game/internal/player"
	"voxel-game/internal/render"
	"voxel-game/internal/ui"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	windowWidth  = 1280
	windowHeight = 720
	windowTitle  = "Voxel Game"
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
	glfw.SwapInterval(1) // 1 = VSync on, 0 = VSync off

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

	// Initialize camera
	cam := camera.NewCamera(windowWidth, windowHeight)

	// Initialize renderer
	renderer, err := render.NewRenderer()
	if err != nil {
		log.Fatalln("failed to create renderer:", err)
	}

	// Initialize UI renderer
	uiRenderer, err := ui.NewUIRenderer(windowWidth, windowHeight)
	if err != nil {
		log.Fatalln("failed to create UI renderer:", err)
	}
	defer uiRenderer.Cleanup()

	// Add UI elements
	crosshair := ui.NewCrosshair(windowWidth, windowHeight)
	if err := uiRenderer.AddElement(crosshair); err != nil {
		log.Fatalln("failed to add crosshair:", err)
	}

	hotbar := ui.NewHotbar(windowWidth, windowHeight)
	if err := uiRenderer.AddElement(hotbar); err != nil {
		log.Fatalln("failed to add hotbar:", err)
	}

	// Window resize callback
	window.SetFramebufferSizeCallback(func(w *glfw.Window, width, height int) {
		gl.Viewport(0, 0, int32(width), int32(height))
		cam.SetSize(width, height)
		uiRenderer.Resize(width, height)

		// Update UI elements with new size
		screenSize := &ui.ScreenSize{Width: width, Height: height}
		crosshair.Update(screenSize)
		hotbar.Update(screenSize)
	})

	// Initialize world
	gameWorld := world.NewWorld()

	// Initialize player
	p := player.NewPlayer(cam, gameWorld)

	wireframeMode := false

	// Initialize input manager
	inputMgr := input.NewInputManager(window, cam, p, &wireframeMode)

	// Debug
	fmt.Printf("BlockAir = %d\n", world.BlockAir)
	fmt.Printf("BlockDirt = %d\n", world.BlockDirt)
	fmt.Printf("BlockGrass = %d\n", world.BlockGrass)
	fmt.Printf("BlockStone = %d\n", world.BlockStone)
	fmt.Printf("Initial selected block = %d\n", inputMgr.GetSelectedBlock())

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
	chunkUpdateInterval := 0.5 // Update chunks every 0.5 seconds

	// Track selected block for hotbar
	lastSelectedBlock := world.BlockAir

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
			lastChunkUpdate = currentTime
		}

		// Update hotbar if selected block changed
		selectedBlock := inputMgr.GetSelectedBlock()
		if selectedBlock != lastSelectedBlock {
			hotbar.Update(selectedBlock)
			lastSelectedBlock = selectedBlock
		}

		// Clear screen
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render world
		renderer.RenderWorld(gameWorld, cam)

		// Render UI
		uiRenderer.Render()

		// Swap buffers and poll events
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
