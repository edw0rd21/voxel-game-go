# Simple Voxel Game (Minecraft Clone)

A simple Minecraft-like voxel game built in Go with OpenGL 4.1.

## Features

- **3D Voxel World:** Infinite world generation with dynamic chunk loading/unloading.
- **Advanced Terrain:** Multi-octave noise generation (Continental, Erosion, Detail layers).
- **Physics Engine:** AABB collision detection, gravity, and raycasting for block interaction.
- **User Interface (UI):**
  - Custom UI rendering engine with texture support.
  - TrueType Font (TTF) rendering with dynamic texture atlases.
  - Interactive Hotbar with block selection.
  - Real-time FPS Counter.
  - **Resolution Independence:** UI scales correctly on High-DPI and 4K monitors.
- **Camera:** First-person camera with smooth view bobbing and mouse look.

### Performance Optimizations:
- **Face Culling:** Hidden block faces are removed from the mesh.
- **Frustum Culling:** Chunks outside the camera's view are not rendered.
- **Chunk Throttling:** Updates are staggered to prevent frame drops.
- **Batch Rendering:** UI elements are batched to minimize draw calls.
- **Directional Shading:** Face-dependent lighting for depth perception.

## Controls

- **WASD** - Move around
- **Mouse** - Look around
- **Space** - Jump
- **Left Click** - Break block
- **Right Click** - Place block
- **1, 2, 3** - Select block type (1=Dirt, 2=Grass, 3=Stone)
- **V** - Toggle creative flying mode (fly but still collide with blocks)
- **N** - Toggle NoClip mode (fly through everything, no collision)
- **F** - Toggle wireframe mode (see mesh optimization)
- **Tab** - Toggle cursor lock (free cursor vs camera control)
- **ESC** - Exit game

## Prerequisites (macOS)

You need to install the following dependencies:

### 1. Install Homebrew (if not already installed)
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### 2. Install Go
```bash
brew install go
```

### 3. Install GLFW dependencies
```bash
brew install glfw
```

### 4. Install pkg-config
```bash
brew install pkg-config
```

## Building and Running

### 1. Navigate to the project directory
```bash
cd voxel-game
```

### 2. Download Go dependencies
```bash
go mod tidy
```

### 3. Build the game
```bash
go build -o game ./cmd/game
```

### 4. Run the game
```bash
./game
```

## Troubleshooting

### "Package glfw was not found" error
Make sure GLFW is installed via Homebrew:
```bash
brew install glfw
pkg-config --modversion glfw3
```

### "fatal error: GL/glew.h: No such file or directory"
This shouldn't happen on macOS as we're using the native OpenGL framework, but if you encounter GL-related errors:
```bash
brew install glew
```

### "ld: library not found"
Make sure pkg-config is installed and can find GLFW:
```bash
brew install pkg-config
pkg-config --libs glfw3
```

### Black screen or no rendering
- Check that your Mac supports OpenGL 4.1 (most Macs from 2012+ do)
- Try updating your graphics drivers
- Check the console output for OpenGL version info

### Performance issues
- If you are not consistently getting 60+ FPS on modern hardware, try the following:

#### Quick fixes:

- Reduce render distance in internal/world/world.go:
```bash
RenderDistance = 2  // Change from 3 to 2
```

- Increase chunk update interval in cmd/game/main.go:
```bash
chunkUpdateInterval := 1.0  // Change from 0.5 to 1.0
```

- Disable VSync for unlimited FPS (more GPU usage):

```bash
glfw.SwapInterval(0)  // Change from 1 to 0
```

## Project Structure

```
voxel-game/
├── assets/                 # Game assets
│   └── fonts/             # .ttf font files
├── cmd/
│   └── game/
│       └── main.go         # Entry point & Game Loop
├── internal/
│   ├── camera/
│   │   └── camera.go       # Camera logic
│   ├── input/
│   │   └── input.go        # Input manager
│   ├── player/
│   │   └── player.go       # Player physics & state
│   ├── render/
│   │   ├── renderer.go     # World Renderer
│   │   └── shaders/        # World Shaders (vertex/fragment)
│   ├── ui/                 # UI System
│   │   ├── font.go         # Font Loader & Atlas Generator
│   │   ├── text.go         # Text Component
│   │   ├── ui.go           # UI Renderer (Ortho projection)
│   │   └── shaders/        # UI Shaders (supports text & color)
│   └── world/
│       └── world.go        # Chunk management & generation
├── go.mod
└── README.md
```

## How It Works

### The UI System
The UI uses a separate rendering pipeline with an Orthographic Projection.

- Uber-Shader: A single shader handles both solid-color shapes (like the crosshair) and textured elements (like text) by using a default white texture for non-textured objects.

- Font Rendering: We use freetype to generate a texture atlas from .ttf files at runtime. Glyphs are batched into a single mesh for high performance.

- Virtual Resolution: The UI uses a reference height (720p). On high-DPI or 4K screens, the projection matrix automatically scales elements to maintain consistent size and readability.

### Chunk System
The world is divided into 16x64x16 chunks. Each chunk generates its own mesh. Only visible block faces are rendered (faces adjacent to other solid blocks are culled).

### Infinite World
- Chunks dynamically load as you explore.
- Chunks outside render distance automatically unload (triggering OpenGL buffer cleanup).
- Generation is deterministic based on a seed.

### Terrain Generation
Advanced multi-octave noise-based generation:
1.  **Continental Layer:** Large-scale landmasses and oceans.
2.  **Erosion Layer:** Medium hills and valleys.
3.  **Detail Layer:** Surface variation.

## Future Improvements

- [ ] Texture mapping for world blocks (currently only UI supports textures)
- [ ] Save/Load world system
- [ ] Water physics and transparency
- [ ] Day/Night cycle
- [ ] Sound effects
- [ ] Inventory system

## Completed Features

- [x] Infinite procedural terrain
- [x] Dynamic chunk loading
- [x] Physics-based movement (Gravity, Collision)
- [x] View Bobbing
- [x] **UI System & Text Rendering**
- [x] **Resolution Independence**
- [x] Debug/God Mode

## License

This is a simple educational project. Feel free to use and modify as you wish!