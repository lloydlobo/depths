package player

import (
	"bytes"
	"cmp"
	"example/depths/internal/common"
	"example/depths/internal/floor"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BoneType uint8

const (
	BoneSocketHat   BoneType = 0
	BoneSocketHandR BoneType = 1
	BoneSocketHandL BoneType = 2
	MaxBoneSockets  BoneType = 3
)

type Player struct {
	Position              rl.Vector3
	Size                  rl.Vector3
	Rotation              int32
	BoundingBox           rl.BoundingBox
	Collisions            rl.Quaternion
	IsPlayerWallCollision bool
}

type ActionType int32

const (
	Idle ActionType = iota
	IdleSway
	Walk
	Mine
)

var (
	action ActionType = Idle
)
var (
	CharacterModel       rl.Model
	CharacterAngle       int32
	EquippedModels       [MaxBoneSockets]rl.Model
	IsShowEquippedModels [MaxBoneSockets]bool
	PlayerCol            = rl.RayWhite
)

var (
	playerModel rl.Model

	modelAnimations  []rl.ModelAnimation
	animsCount       = uint(0)
	animIndex        = uint(0)
	animCurrentFrame = uint(0)
	boneSocketsIndex [MaxBoneSockets]int
)

var (
	anim            rl.ModelAnimation
	characterRotate rl.Quaternion
)

func NewPlayer(camera rl.Camera3D) Player {
	player := Player{
		Position:   camera.Target,
		Size:       cmp.Or(rl.NewVector3(.5, 1.-.5, .5), rl.NewVector3(1, 2, 1)),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}
	player.BoundingBox = common.GetBoundingBoxFromPositionSizeV(camera.Target, player.Size)

	return player
}

func InitPlayer(player *Player, camera rl.Camera3D) {
	*player = NewPlayer(camera)
	// FIXME: Remove this or bring the one from NewPlayer here
}

func SetupPlayerModel() {
	playerModel = common.Model.OBJ.CharacterHuman
	rl.SetMaterialTexture(playerModel.Materials, rl.MapDiffuse, common.Model.OBJ.Colormap)

	// INIT

	// Load gltf model
	// See https://www.raylib.com/examples/models/loader.html?name=models_bone_socket
	// See https://github.com/raysan5/raylib/tree/master/examples/models/resources/models/gltf
	CharacterModel = rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman.glb"))
	EquippedModels = [MaxBoneSockets]rl.Model{
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_hat.glb")),    // Index for the hat model is the same as BONE_SOCKET_HAT
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_sword.glb")),  // Index for the sword model is the same as BONE_SOCKET_HAND_R
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_shield.glb")), // Index for the shield model is the same as BONE_SOCKET_HAND_L
	}
	IsShowEquippedModels = [MaxBoneSockets]bool{true, true, true}

	// Load gltf model animations
	animIndex = 0
	animCurrentFrame = 0
	modelAnimations = rl.LoadModelAnimations(filepath.Join("res", "model", "gltf", "greenman.glb"))
	animsCount = uint(len(modelAnimations))

	// Indices for bones for sockets
	boneSocketsIndex = [MaxBoneSockets]int{-1, -1, -1}

	// See https://stackoverflow.com/questions/28848187/how-to-convert-int8-to-string
	B2S := func(bs []int8) string {
		b := make([]byte, len(bs))
		for i, v := range bs {
			b[i] = byte(v)
		}
		var sb strings.Builder
		sb.Write(b)
		return sb.String()
	}

	// Search bones for sockets in -> [root,body_low,body_up,socket_hat,hand_L,hand_R,hip_L,leg_L,hip_R,leg_R,socket_hand_L,socket_hand_R]
	for i := range CharacterModel.BoneCount {
		var buf [32]int8 = CharacterModel.GetBones()[i].Name
		var name string = B2S(buf[:])

		// FIXME: String comparison not work with == operator
		if bytes.Equal([]byte(name), []byte("socket_hat")) ||
			(!strings.EqualFold(name, "socket_hat") &&
				(strings.Contains(name, "socket") && strings.Contains(name, "hat"))) {
			boneSocketsIndex[BoneSocketHat] = int(i)
			continue
		}
		if bytes.Equal([]byte(name), []byte("socket_hand_R")) ||
			(!strings.EqualFold(name, "socket_hand_R") &&
				(strings.Contains(name, "socket") && strings.Contains(name, "hand") && strings.Contains(name, "R"))) {
			boneSocketsIndex[BoneSocketHandR] = int(i)
			continue
		}
		if bytes.Equal([]byte(name), []byte("socket_hand_L")) ||
			(!strings.EqualFold(name, "socket_hand_L") &&
				(strings.Contains(name, "socket") && strings.Contains(name, "hand") && strings.Contains(name, "L"))) {
			boneSocketsIndex[BoneSocketHandL] = int(i)
			continue
		}
	}

	if got, want := boneSocketsIndex[:], [3]int{3, 11, 10}; !slices.Equal(got[:], want[:]) { // boneSocketIndex => initial [-1,-1,-1] => want [3,11,10]
		panic(fmt.Sprintln("NewPlayer: boneSocketIndex", "got", got, "want", want))
	}
}

func (p *Player) Update(camera rl.Camera3D, flr floor.Floor) {
	if rl.IsKeyDown(rl.KeyW) ||
		rl.IsKeyDown(rl.KeyA) ||
		rl.IsKeyDown(rl.KeyS) ||
		rl.IsKeyDown(rl.KeyD) {
		action = Walk
	} else {
		action = IdleSway
	}

	// Overide movement actions
	if rl.IsKeyDown(rl.KeySpace) {
		action = Mine
	}

	// Rotate character
	if int32(p.Rotation) != characterAngle {
		if p.Rotation < 0 {
			characterAngle = (int32(p.Rotation) + 1*0) % 360
		} else {
			characterAngle = (360 + int32(p.Rotation) - 1*0) % 360
		}
	}
	if rl.IsKeyDown(rl.KeyH) {
		characterAngle = (characterAngle + 1) % 360
	} else if rl.IsKeyDown(rl.KeyL) {
		characterAngle = (360 + characterAngle - 1) % 360
	}
	// Select current animation
	if true {
		switch action {
		case Idle:
			animIndex = 0
		case IdleSway:
			animIndex = 1
		case Walk:
			animIndex = 2
		case Mine:
			animIndex = 3
		default:
			panic(fmt.Sprintf("unexpected player.ActionType: %#v", action))
		}
		if animIndex >= animsCount {
			panic(fmt.Sprintf("unexpected player.animIndex: %#v", animIndex))
		}
	} else { // DEBUG
		if rl.IsKeyPressed(rl.KeyT) {
			if animsCount > 0 {
				animIndex = (animIndex + 1) % animsCount
			}
		} else if rl.IsKeyPressed(rl.KeyG) {
			if animsCount > 0 {
				animIndex = (animIndex + animsCount - 1) % animsCount
			}
		}
	}

	// Toggle shown of equip
	if rl.IsKeyPressed(rl.KeyOne) {
		IsShowEquippedModels[BoneSocketHat] = !IsShowEquippedModels[BoneSocketHat]
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		IsShowEquippedModels[BoneSocketHandR] = !IsShowEquippedModels[BoneSocketHandR]
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		IsShowEquippedModels[BoneSocketHandL] = !IsShowEquippedModels[BoneSocketHandL]
	}

	// Update model animation
	anim = modelAnimations[animIndex]
	if anim.FrameCount > 0 {
		animCurrentFrame = (animCurrentFrame + 1) % uint(anim.FrameCount)
		rl.UpdateModelAnimation(CharacterModel, anim, int32(animCurrentFrame))
	}

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
	if false {
		rl.DrawModelEx(playerModel,
			rl.NewVector3(p.Position.X, p.Position.Y-p.Size.Y/2, p.Position.Z),
			rl.NewVector3(0, 1, 0), 0.0,
			rl.NewVector3(1., common.InvPhi, 1.), rl.White)
		rl.DrawCapsule(
			rl.Vector3Add(p.Position, rl.NewVector3(0, p.Size.Y/8, 0)),
			rl.Vector3Add(p.Position, rl.NewVector3(0, -p.Size.Y/4, 0)),
			p.Size.X/2, 8, 8, PlayerCol)
	} else {
		// Draw character and equipments
		rl.PushMatrix()

		const scaleToReduceBy = 3.
		const cameraTargetPlayerCenterYOffset = .5

		posX := p.Position.X
		posY := (p.Position.Y - cameraTargetPlayerCenterYOffset)
		posZ := p.Position.Z

		// NOTE: Transformation is applied in inverse order (scale -> rotate -> translate)
		if false {
			rl.Translatef(2.0, 0.0, 0.0)
		}
		if false {
			rl.Rotatef(45, 0, 1, 0)
		}
		if true {
			rl.Scalef(1.*1./scaleToReduceBy, 1.*1./scaleToReduceBy, 1.*1./scaleToReduceBy)
			posX *= scaleToReduceBy
			posY *= scaleToReduceBy
			posZ *= scaleToReduceBy
		}

		// Draw character
		characterRotate = rl.QuaternionFromAxisAngle(rl.NewVector3(0.0, 1.0, 0.0), float32(CharacterAngle)*rl.Deg2rad)
		CharacterModel.Transform = rl.MatrixMultiply(rl.QuaternionToMatrix(characterRotate), rl.MatrixTranslate(posX, posY, posZ))
		rl.UpdateModelAnimation(CharacterModel, anim, int32(animCurrentFrame))
		rl.DrawMesh(CharacterModel.GetMeshes()[0],
			CharacterModel.GetMaterials()[1], CharacterModel.Transform)

		// Draw equipments (hat, sword, shield)
		for i := range MaxBoneSockets {
			if !IsShowEquippedModels[i] {
				continue
			}
			if anim.FramePoses == nil || CharacterModel.BindPose == nil {
				continue
			}

			var transform rl.Transform = anim.GetFramePose(int(animCurrentFrame), boneSocketsIndex[i])
			var inRotation rl.Quaternion = CharacterModel.GetBindPose()[boneSocketsIndex[i]].Rotation
			var outRotation rl.Quaternion = transform.Rotation

			// Calculate socket rotation (angle between bone in initial pose and same bone in current animation frame)
			var rotate rl.Quaternion = rl.QuaternionMultiply(outRotation, rl.QuaternionInvert(inRotation))
			var matrixTransform rl.Matrix = rl.QuaternionToMatrix(rotate)
			// Translate socket to its position in the current animation
			matrixTransform = rl.MatrixMultiply(matrixTransform, rl.MatrixTranslate(transform.Translation.X, transform.Translation.Y, transform.Translation.Z))
			// Transform the socket using the transform of the character (angle and translate)
			matrixTransform = rl.MatrixMultiply(matrixTransform, CharacterModel.Transform)

			// Draw mesh at socket position with socket angle rotation
			rl.DrawMesh(EquippedModels[i].GetMeshes()[0], EquippedModels[i].GetMaterials()[1], matrixTransform)
		}

		rl.PopMatrix()
	}

	// Debug
	if true {
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
		if p.IsPlayerWallCollision {
			rl.DrawBoundingBox(p.BoundingBox, rl.Red)
		}
		if true {
			common.DrawXYZOrbitV(p.Position, 1./common.Phi)
		}
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
