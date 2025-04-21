package player

import (
	"cmp"
	"example/depths/internal/common"
	"example/depths/internal/floor"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Player struct {
	Position              rl.Vector3
	Size                  rl.Vector3
	BoundingBox           rl.BoundingBox
	Collisions            rl.Quaternion
	IsPlayerWallCollision bool
}

var (
	PlayerCol = cmp.Or(rl.White, rl.ColorLerp(rl.Black, rl.DarkGray, .1))
)

var (
	playerModel rl.Model
)

func NewPlayer(camera rl.Camera3D) Player {
	out := Player{
		Position:   camera.Target,
		Size:       cmp.Or(rl.NewVector3(.5, 1.-.5, .5), rl.NewVector3(1, 2, 1)),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}
	out.BoundingBox = common.GetBoundingBoxFromPositionSizeV(camera.Target, out.Size)
	return out
}

func InitPlayer(player *Player, camera rl.Camera3D) {
	*player = NewPlayer(camera)
	playerModel = common.Model.OBJ.CharacterHuman
	rl.SetMaterialTexture(playerModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)
}

func (p *Player) Update(camera rl.Camera3D, flr floor.Floor) {
	// Project the player as the camera target
	p.Position = camera.Target

	p.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(p.Position.X-p.Size.X/2, p.Position.Y-p.Size.Y/2, p.Position.Z-p.Size.Z/2),
		rl.NewVector3(p.Position.X+p.Size.X/2, p.Position.Y+p.Size.Y/2, p.Position.Z+p.Size.Z/2))

	// Wall collisions
	if p.BoundingBox.Min.X <= flr.BoundingBox.Min.X {
		p.IsPlayerWallCollision = true
		p.Collisions.X = -1
	}
	if p.BoundingBox.Max.X >= flr.BoundingBox.Max.X {
		p.IsPlayerWallCollision = true
		p.Collisions.X = 1
	}
	if p.BoundingBox.Min.Z <= flr.BoundingBox.Min.Z {
		p.IsPlayerWallCollision = true
		p.Collisions.Z = -1
	}
	if p.BoundingBox.Max.Z >= flr.BoundingBox.Max.Z {
		p.IsPlayerWallCollision = true
		p.Collisions.Z = 1
	}

	// Floor collisions
	if p.BoundingBox.Min.Y <= flr.BoundingBox.Min.Y {
		p.Collisions.Y = 1 // Player head below floor
	}
	if p.BoundingBox.Max.Y >= flr.BoundingBox.Min.Y { // On floor
		p.Collisions.W = -1 // Allow walking freely
	}
}

func (p Player) Draw() {
	rl.DrawModelEx(playerModel,
		rl.NewVector3(p.Position.X, p.Position.Y-p.Size.Y/2, p.Position.Z),
		rl.NewVector3(0, 1, 0), 0.0,
		rl.NewVector3(1., common.InvPhi, 1.), rl.White)
	rl.DrawCapsule(
		rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/8, 0)),
		rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
		p.Size.X/2, 8, 8, rl.Fade(PlayerCol, .5))

	// Debug
	if true {
		if p.IsPlayerWallCollision {
			rl.DrawBoundingBox(p.BoundingBox, rl.Red)
		}
		size := rl.Vector3Scale(p.Size, .5)

		if p.Collisions.X != 0 {
			pos := p.Position
			pos.X += p.Collisions.X * p.Size.X / 2
			rl.DrawCubeV(pos, size, common.XAxisColor)
		}
		if p.Collisions.Y != 0 {
			pos := p.Position
			pos.Y += p.Collisions.Y * p.Size.Y / 2
			rl.DrawCubeV(pos, size, common.YAxisColor)
		}
		if p.Collisions.Z != 0 {
			pos := p.Position
			pos.Z += p.Collisions.Z * p.Size.Z / 2
			rl.DrawCubeV(pos, size, common.ZAxisColor)
		}
		if p.Collisions.W != 0 { // Floor
			pos := p.Position
			pos.Y += p.Collisions.W * p.Size.Y / 2
			rl.DrawCubeV(pos, size, common.YAxisColor)
		}

		common.DrawXYZOrbitV(p.Position, 1./common.Phi)
	}
}

// FIXME: Camera gets stuck if player keeps moving into the box. Maybe lerp or
// free camera if "distance to the box is less" or touching.
func RevertPlayerAndCameraPositions(
	srcPlayer Player, dstPlayer *Player,
	srcCamera rl.Camera3D, dstCamera *rl.Camera3D,
) {
	dstPlayer.Position = srcPlayer.Position
	dstPlayer.BoundingBox = rl.NewBoundingBox(
		rl.NewVector3(dstPlayer.Position.X-dstPlayer.Size.X/2,
			dstPlayer.Position.Y-dstPlayer.Size.Y/2,
			dstPlayer.Position.Z-dstPlayer.Size.Z/2),
		rl.NewVector3(dstPlayer.Position.X+dstPlayer.Size.X/2,
			dstPlayer.Position.Y+dstPlayer.Size.Y/2,
			dstPlayer.Position.Z+dstPlayer.Size.Z/2))
	dstCamera.Target = srcCamera.Target
	dstCamera.Position = srcCamera.Position
}
