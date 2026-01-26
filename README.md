# Simple Voxel Game (Minecraft Clone)

A simple Minecraft-like voxel game built in Go with OpenGL.

## Features

- ✅ 3D voxel-based world with chunk system
- ✅ Procedural terrain generation
- ✅ First-person camera with mouse look
- ✅ WASD movement with physics (gravity, collision)
- ✅ Block placement and destruction
- ✅ Multiple block types (grass, dirt, stone)
- ✅ Optimized rendering (only visible faces)

## Controls

- **WASD** - Move around
- **Mouse** - Look around
- **Space** - Jump
- **Left Click** - Break block
- **Right Click** - Place block
- **1, 2, 3** - Select block type (1=Dirt, 2=Grass, 3=Stone)
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
go mod download
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
- The game generates chunks around spawn, which might take a moment
- Try reducing the render distance by modifying the chunk generation range in `world.go`

## Project Structure

```
voxel-game/
├── cmd/
│   └── game/
│       └── main.go                 # Entry point
├── internal/
│   ├── camera/
│   │   └── camera.go              # Camera controls
│   ├── input/
│   │   └── input.go               # Input handling
│   ├── player/
│   │   └── player.go              # Player physics & raycasting
│   ├── render/
│   │   ├── renderer.go            # OpenGL rendering
│   │   └── shaders/
│   │       ├── vertex.glsl        # Vertex shader
│   │       └── fragment.glsl      # Fragment shader
│   └── world/
│       └── world.go               # World & chunk management
├── go.mod
└── README.md
```

## How It Works

### Chunk System
The world is divided into 16x64x16 chunks. Each chunk contains blocks and generates its own mesh. Only visible block faces are rendered (faces adjacent to other solid blocks are culled).

### Terrain Generation
Simple procedural generation using a sine/cosine function for height variation. Each chunk generates terrain based on world coordinates to ensure continuity.

### Mesh Generation
For each chunk, the system:
1. Iterates through all blocks
2. Checks each face of solid blocks
3. Only adds faces that are exposed (adjacent to air or chunk boundary)
4. Combines all faces into a single mesh per chunk

### Physics
- Gravity constantly pulls the player down
- AABB (Axis-Aligned Bounding Box) collision detection
- Separate collision checks for X, Y, Z axes
- Player is 0.6 blocks wide and 1.8 blocks tall

### Raycasting
When placing/breaking blocks:
1. Cast a ray from camera position in the look direction
2. Check each point along the ray (0.1 block intervals)
3. When a solid block is hit, determine which face was hit
4. For breaking: remove the hit block
5. For placing: add a block on the hit face

## Future Improvements

Some ideas for extending this project:
- Texture mapping instead of solid colors
- More block types
- Inventory system
- Saving/loading worlds
- Water and transparency
- Day/night cycle with lighting
- Particles and sound effects
- Multiplayer support

## License

This is a simple educational project. Feel free to use and modify as you wish!
