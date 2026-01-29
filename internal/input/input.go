package input

import (
	"fmt"
	"voxel-game/internal/camera"
	"voxel-game/internal/player"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type InputManager struct {
	window *glfw.Window
	camera *camera.Camera
	player *player.Player

	firstMouse bool
	lastX      float64
	lastY      float64

	selectedBlock world.BlockType
	cursorLocked  bool

	//Debug State
	debugMode bool
	flySpeed  float32
	wireframe *bool
}

func NewInputManager(window *glfw.Window, cam *camera.Camera, p *player.Player, wireframe *bool) *InputManager {
	im := &InputManager{
		window:        window,
		camera:        cam,
		player:        p,
		firstMouse:    true,
		selectedBlock: world.BlockDirt,
		cursorLocked:  true,
		wireframe:     wireframe,
		flySpeed:      20.0,
	}

	// Set up callbacks
	window.SetCursorPosCallback(im.mouseCallback)
	window.SetMouseButtonCallback(im.mouseButtonCallback)
	window.SetKeyCallback(im.keyCallback)

	return im
}

func (im *InputManager) IsDebugMode() bool {
	return im.debugMode
}

func (im *InputManager) GetSelectedBlock() world.BlockType {
	return im.selectedBlock
}

func (im *InputManager) Update(deltaTime float32) {
	// Close window
	if im.window.GetKey(glfw.KeyEscape) == glfw.Press {
		im.window.SetShouldClose(true)
	}

	// STATE MACHINE: Switch controls based on mode
	if im.debugMode {
		im.updateDebugCamera(deltaTime)
	} else {
		im.updatePlayer(deltaTime)
	}
}

func (im *InputManager) updatePlayer(deltaTime float32) {
	var moveDir mgl32.Vec3

	// Standard WASD
	if im.window.GetKey(glfw.KeyW) == glfw.Press {
		forward := im.camera.Front
		forward[1] = 0 // Keep player stuck to ground plane
		forward = forward.Normalize()
		moveDir = moveDir.Add(forward)
	}
	if im.window.GetKey(glfw.KeyS) == glfw.Press {
		forward := im.camera.Front
		forward[1] = 0
		forward = forward.Normalize()
		moveDir = moveDir.Sub(forward)
	}
	if im.window.GetKey(glfw.KeyA) == glfw.Press {
		moveDir = moveDir.Sub(im.camera.Right)
	}
	if im.window.GetKey(glfw.KeyD) == glfw.Press {
		moveDir = moveDir.Add(im.camera.Right)
	}

	// Apply movement
	if moveDir.Len() > 0 {
		moveDir = moveDir.Normalize()
		im.player.Move(moveDir, deltaTime)
	}

	// Player Actions
	if im.window.GetKey(glfw.KeySpace) == glfw.Press {
		im.player.Jump()
	}
}

func (im *InputManager) updateDebugCamera(deltaTime float32) {
	// Calculate Speed
	currentSpeed := im.flySpeed
	if im.window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		currentSpeed *= 3.0 // Sprint (Fast Fly)
	}
	if im.window.GetKey(glfw.KeyLeftControl) == glfw.Press {
		currentSpeed *= 0.1 // Precision Mode (Slow Fly)
	}

	// Free Fly Movement (Moves Camera.Position directly)
	// W/S = Forward/Backward (in looking direction)
	if im.window.GetKey(glfw.KeyW) == glfw.Press {
		im.camera.Position = im.camera.Position.Add(im.camera.Front.Mul(currentSpeed * deltaTime))
	}
	if im.window.GetKey(glfw.KeyS) == glfw.Press {
		im.camera.Position = im.camera.Position.Sub(im.camera.Front.Mul(currentSpeed * deltaTime))
	}
	// A/D = Strafe Left/Right
	if im.window.GetKey(glfw.KeyA) == glfw.Press {
		im.camera.Position = im.camera.Position.Sub(im.camera.Right.Mul(currentSpeed * deltaTime))
	}
	if im.window.GetKey(glfw.KeyD) == glfw.Press {
		im.camera.Position = im.camera.Position.Add(im.camera.Right.Mul(currentSpeed * deltaTime))
	}
	// Space/Alt = Up/Down (Absolute World Up)
	if im.window.GetKey(glfw.KeySpace) == glfw.Press {
		im.camera.Position = im.camera.Position.Add(im.camera.WorldUp.Mul(currentSpeed * deltaTime))
	}
	if im.window.GetKey(glfw.KeyLeftAlt) == glfw.Press {
		im.camera.Position = im.camera.Position.Sub(im.camera.WorldUp.Mul(currentSpeed * deltaTime))
	}
}

func (im *InputManager) mouseCallback(w *glfw.Window, xpos, ypos float64) {
	if !im.cursorLocked {
		return
	}
	if im.firstMouse {
		im.lastX = xpos
		im.lastY = ypos
		im.firstMouse = false
	}

	xoffset := xpos - im.lastX
	yoffset := im.lastY - ypos // Reversed since y-coordinates go from bottom to top

	im.lastX = xpos
	im.lastY = ypos

	im.camera.ProcessMouseMovement(float32(xoffset), float32(yoffset))
}

func (im *InputManager) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		if button == glfw.MouseButtonLeft {
			// Break block
			im.player.BreakBlock()
		} else if button == glfw.MouseButtonRight {
			// Place block
			im.player.PlaceBlock(im.selectedBlock)
		}
	}
}

func (im *InputManager) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		// Number keys to select block type
		switch key {
		case glfw.Key1:
			im.selectedBlock = world.BlockDirt
		case glfw.Key2:
			im.selectedBlock = world.BlockGrass
		case glfw.Key3:
			im.selectedBlock = world.BlockStone
		case glfw.KeyTab:
			im.cursorLocked = !im.cursorLocked
			if im.cursorLocked {
				w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			} else {
				w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			}
		case glfw.KeyG:
			im.debugMode = !im.debugMode
			fmt.Printf("Debug Mode: %v\n", im.debugMode)
			// Unfreeze frustum when exiting debug mode so we don't get stuck with a weird view
			if !im.debugMode {
				im.camera.FrustumFrozen = false
				// Force wireframe off when leaving debug mode
				if *im.wireframe {
					*im.wireframe = false
					gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
				}
			}
		case glfw.KeyF:
			if im.debugMode {
				*im.wireframe = !*im.wireframe
				if *im.wireframe {
					gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
				} else {
					gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
				}
				fmt.Printf("Wireframe: %v\n", *im.wireframe)
			}

		case glfw.KeyP:
			// Toggle Frustum Freeze (Only works in Debug Mode)
			if im.debugMode {
				im.camera.FrustumFrozen = !im.camera.FrustumFrozen
				fmt.Printf("Frustum Frozen: %v\n", im.camera.FrustumFrozen)
			}
		}
	}
}
