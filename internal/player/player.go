package player

import (
	"math"

	"voxel-game/internal/camera"
	"voxel-game/internal/world"

	"github.com/go-gl/mathgl/mgl32"
)

type TargetBlock struct {
	Hit  bool
	Pos  mgl32.Vec3
	Face int
}

type Player struct {
	camera   *camera.Camera
	world    *world.World
	speed    float32
	velocity mgl32.Vec3

	grounded bool
	width    float32
	height   float32

	target TargetBlock
}

func NewPlayer(cam *camera.Camera, w *world.World) *Player {
	return &Player{
		camera: cam,
		world:  w,
		width:  0.6,
		height: 1.8,
		speed:  1.2,
	}
}

func (p *Player) Update(deltaTime float32) {
	// Apply gravity
	if !p.grounded {
		p.velocity[1] -= 20.0 * deltaTime
	}

	// Apply velocity
	newPos := p.camera.Position.Add(p.velocity.Mul(deltaTime))

	// Collision detection
	newPos = p.handleCollision(newPos)

	p.camera.Position = newPos

	// Check if grounded
	p.grounded = p.isGrounded()

	// Damping
	p.velocity = p.velocity.Mul(0.8)

	p.updateTarget()
}

func (p *Player) updateTarget() {
	hit, x, y, z, face := p.Raycast(5.0)
	if hit {
		p.target = TargetBlock{
			Hit:  true,
			Pos:  mgl32.Vec3{float32(x), float32(y), float32(z)},
			Face: face,
		}
	} else {
		p.target = TargetBlock{}
	}
}

func (p *Player) TargetBlock() TargetBlock {
	return p.target
}

func (p *Player) Move(direction mgl32.Vec3) {
	p.velocity = p.velocity.Add(direction.Mul(p.speed))
}

func (p *Player) Jump() {
	if p.grounded {
		p.velocity[1] = 8.0
	}
}

func (p *Player) handleCollision(newPos mgl32.Vec3) mgl32.Vec3 {
	// Simple AABB collision
	minX := int(math.Floor(float64(newPos[0] - p.width/2)))
	maxX := int(math.Floor(float64(newPos[0] + p.width/2)))
	minY := int(math.Floor(float64(newPos[1])))
	maxY := int(math.Floor(float64(newPos[1] + p.height)))
	minZ := int(math.Floor(float64(newPos[2] - p.width/2)))
	maxZ := int(math.Floor(float64(newPos[2] + p.width/2)))

	// Check X axis
	for y := minY; y <= maxY; y++ {
		for z := minZ; z <= maxZ; z++ {
			if p.world.GetBlock(minX, y, z) != world.BlockAir {
				newPos[0] = p.camera.Position[0]
				p.velocity[0] = 0
				break
			}
			if p.world.GetBlock(maxX, y, z) != world.BlockAir {
				newPos[0] = p.camera.Position[0]
				p.velocity[0] = 0
				break
			}
		}
	}

	// Check Y axis
	minX = int(math.Floor(float64(newPos[0] - p.width/2)))
	maxX = int(math.Floor(float64(newPos[0] + p.width/2)))
	minZ = int(math.Floor(float64(newPos[2] - p.width/2)))
	maxZ = int(math.Floor(float64(newPos[2] + p.width/2)))

	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			if p.world.GetBlock(x, minY, z) != world.BlockAir {
				newPos[1] = p.camera.Position[1]
				p.velocity[1] = 0
				break
			}
			if p.world.GetBlock(x, maxY, z) != world.BlockAir {
				newPos[1] = p.camera.Position[1]
				p.velocity[1] = 0
				break
			}
		}
	}

	// Check Z axis
	minX = int(math.Floor(float64(newPos[0] - p.width/2)))
	maxX = int(math.Floor(float64(newPos[0] + p.width/2)))
	minY = int(math.Floor(float64(newPos[1])))
	maxY = int(math.Floor(float64(newPos[1] + p.height)))

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			if p.world.GetBlock(x, y, minZ) != world.BlockAir {
				newPos[2] = p.camera.Position[2]
				p.velocity[2] = 0
				break
			}
			if p.world.GetBlock(x, y, maxZ) != world.BlockAir {
				newPos[2] = p.camera.Position[2]
				p.velocity[2] = 0
				break
			}
		}
	}

	return newPos
}

func (p *Player) isGrounded() bool {
	minX := int(math.Floor(float64(p.camera.Position[0] - p.width/2)))
	maxX := int(math.Floor(float64(p.camera.Position[0] + p.width/2)))
	minZ := int(math.Floor(float64(p.camera.Position[2] - p.width/2)))
	maxZ := int(math.Floor(float64(p.camera.Position[2] + p.width/2)))
	checkY := int(math.Floor(float64(p.camera.Position[1]))) - 1

	for x := minX; x <= maxX; x++ {
		for z := minZ; z <= maxZ; z++ {
			if p.world.GetBlock(x, checkY, z) != world.BlockAir {
				return true
			}
		}
	}

	return false
}

// Raycast to find the block the player is looking at
func (p *Player) Raycast(maxDistance float32) (hit bool, x, y, z int, face int) {
	pos := p.camera.Position
	dir := p.camera.Front
	step := float32(0.1)

	for dist := float32(0); dist < maxDistance; dist += step {
		checkPos := pos.Add(dir.Mul(dist))
		bx := int(math.Floor(float64(checkPos[0])))
		by := int(math.Floor(float64(checkPos[1])))
		bz := int(math.Floor(float64(checkPos[2])))

		if p.world.GetBlock(bx, by, bz) != world.BlockAir {
			// Determine which face was hit
			prevPos := pos.Add(dir.Mul(dist - step))
			px := int(math.Floor(float64(prevPos[0])))
			py := int(math.Floor(float64(prevPos[1])))
			pz := int(math.Floor(float64(prevPos[2])))

			if bx != px {
				if bx > px {
					face = 3 // -X
				} else {
					face = 2 // +X
				}
			} else if by != py {
				if by > py {
					face = 5 // -Y
				} else {
					face = 4 // +Y
				}
			} else if bz != pz {
				if bz > pz {
					face = 1 // -Z
				} else {
					face = 0 // +Z
				}
			}

			return true, bx, by, bz, face
		}
	}

	return false, 0, 0, 0, 0
}

func (p *Player) BreakBlock() {
	if !p.target.Hit {
		return
	}

	pos := p.target.Pos
	p.world.SetBlock(
		int(pos.X()),
		int(pos.Y()),
		int(pos.Z()),
		world.BlockAir,
	)
}

func (p *Player) PlaceBlock(blockType world.BlockType) {
	if !p.target.Hit {
		return
	}

	x := int(p.target.Pos.X())
	y := int(p.target.Pos.Y())
	z := int(p.target.Pos.Z())

	switch p.target.Face {
	case 0:
		z++
	case 1:
		z--
	case 2:
		x++
	case 3:
		x--
	case 4:
		y++
	case 5:
		y--
	}

	if p.collidesWithPlayer(float32(x), float32(y), float32(z)) {
		return
	}

	p.world.SetBlock(x, y, z, blockType)

}

func (p *Player) collidesWithPlayer(x, y, z float32) bool {
	px := p.camera.Position.X()
	py := p.camera.Position.Y()
	pz := p.camera.Position.Z()

	return mgl32.Abs(px-x) < p.width &&
		py < y+p.height &&
		py+p.height > y &&
		mgl32.Abs(pz-z) < p.width
}
