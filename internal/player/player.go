package player

import (
	"bytes"
	"cmp"
	"fmt"
	"image/color"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"

	"example/depths/internal/common"
	"example/depths/internal/floor"
	"example/depths/internal/util/mathutil"
)

type Player struct {
	Position    rl.Vector3
	Size        rl.Vector3
	Rotation    int32 // FIXME: We are adding 90 degree to adapt to original model's default rotation
	BoundingBox rl.BoundingBox

	Collisions            rl.Quaternion
	IsPlayerWallCollision bool

	Health           float32 // [0..1]
	CargoCapacity    int32   // [0..80]
	MaxCargoCapacity int32   // upgrade lvl01=>80
}

func NewPlayer(camera rl.Camera3D) Player {
	player := Player{
		Position:   camera.Target,
		Size:       cmp.Or(rl.NewVector3(.5, 1.-.5, .5), rl.NewVector3(1, 2, 1)),
		Collisions: rl.NewQuaternion(0, 0, 0, 0),
	}

	player.BoundingBox = common.GetBoundingBoxFromPositionSizeV(camera.Target, player.Size)

	currId := int(common.SavedgameSlotData.CurrentLevelID) // [1..]
	player.MaxCargoCapacity = []int32{80, 86, 92, 96, 108, 116, 128, 136, 146, 156, 186, 206, 216, 236, 266, 296, 306}[currId-1]
	player.CargoCapacity = 0

	player.Health = 1.

	return player
}

// FIXME: Remove this or bring the one from NewPlayer here
func InitPlayer(player *Player, camera rl.Camera3D) {
	*player = NewPlayer(camera)
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

func Action() ActionType {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	return action
}

var (
	playerColor = rl.RayWhite
)

func SetColor(col color.RGBA) {
	playerColor = col // use mutex?
}

type BoneType uint8

const (
	BoneSocketHat   BoneType = 0
	BoneSocketHandR BoneType = 1
	BoneSocketHandL BoneType = 2
	MaxBoneSockets  BoneType = 3
)

var (
	characterModel       rl.Model
	characterAngle       int32
	equippedModels       [MaxBoneSockets]rl.Model
	isShowEquippedModels [MaxBoneSockets]bool
)

var (
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

// FIXME: This has File i/o logic.. Should use resources loaded common to load models apriori
func SetupPlayerModel() {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	// Load gltf model
	// See https://www.raylib.com/examples/models/loader.html?name=models_bone_socket
	// See https://github.com/raysan5/raylib/tree/master/examples/models/resources/models/gltf
	characterModel = rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman.glb"))
	equippedModels = [MaxBoneSockets]rl.Model{
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_hat.glb")),    // Index for the hat model is the same as BONE_SOCKET_HAT
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_sword.glb")),  // Index for the sword model is the same as BONE_SOCKET_HAND_R
		rl.LoadModel(filepath.Join("res", "model", "gltf", "greenman_shield.glb")), // Index for the shield model is the same as BONE_SOCKET_HAND_L
	}

	isShowEquippedModels = [MaxBoneSockets]bool{true, true, true}

	// Load gltf model animations
	animIndex = 0
	animCurrentFrame = 0
	modelAnimations = rl.LoadModelAnimations(filepath.Join("res", "model", "gltf", "greenman.glb"))
	animsCount = uint(len(modelAnimations))

	// Indices for bones for sockets
	boneSocketsIndex = [MaxBoneSockets]int{-1, -1, -1}

	// See https://stackoverflow.com/questions/28848187/how-to-convert-int8-to-string
	byteToString := func(bs []int8) string {
		b := make([]byte, len(bs))
		for i, v := range bs {
			b[i] = byte(v)
		}
		return string(b)
	}

	// Search bones for sockets in -> [root,body_low,body_up,socket_hat,hand_L,hand_R,hip_L,leg_L,hip_R,leg_R,socket_hand_L,socket_hand_R]
	for i := range characterModel.BoneCount {
		var buf [32]int8 = characterModel.GetBones()[i].Name
		var name string = byteToString(buf[:])

		// FIXME: String comparison not work with == operator
		if bytes.Equal([]byte(name), []byte("socket_hat")) ||
			(!strings.EqualFold(name, "socket_hat") && (strings.Contains(name, "socket") && strings.Contains(name, "hat"))) {
			boneSocketsIndex[BoneSocketHat] = int(i)
			continue
		}
		if bytes.Equal([]byte(name), []byte("socket_hand_R")) ||
			(!strings.EqualFold(name, "socket_hand_R") && (strings.Contains(name, "socket") && strings.Contains(name, "hand") && strings.Contains(name, "R"))) {
			boneSocketsIndex[BoneSocketHandR] = int(i)
			continue
		}
		if bytes.Equal([]byte(name), []byte("socket_hand_L")) ||
			(!strings.EqualFold(name, "socket_hand_L") && (strings.Contains(name, "socket") && strings.Contains(name, "hand") && strings.Contains(name, "L"))) {
			boneSocketsIndex[BoneSocketHandL] = int(i)
			continue
		}
	}

	if got, want := boneSocketsIndex[:], [3]int{3, 11, 10}; !slices.Equal(got[:], want[:]) {
		// boneSocketIndex => initial [-1,-1,-1] => want [3,11,10]
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
	if rl.IsKeyDown(rl.KeySpace) || rl.IsMouseButtonDown(rl.MouseLeftButton) {
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
		isShowEquippedModels[BoneSocketHat] = !isShowEquippedModels[BoneSocketHat]
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		isShowEquippedModels[BoneSocketHandR] = !isShowEquippedModels[BoneSocketHandR]
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		isShowEquippedModels[BoneSocketHandL] = !isShowEquippedModels[BoneSocketHandL]
	}

	// Update model animation
	anim = modelAnimations[animIndex]
	if anim.FrameCount > 0 {
		animCurrentFrame = (animCurrentFrame + 1) % uint(anim.FrameCount)
		rl.UpdateModelAnimation(characterModel, anim, int32(animCurrentFrame))
	}

	// Project the player as the camera target
	p.Position = camera.Target
	p.BoundingBox = common.GetBoundingBoxFromPositionSizeV(p.Position, p.Size)

	// Update rotation based on camera forward projection
	startPos := p.Position
	endPos := rl.Vector3Add(p.Position, rl.GetCameraForward(&camera))
	degree := mathutil.Angle2D(startPos.X, startPos.Z, endPos.X, endPos.Z)
	p.Rotation = -90 + int32(degree) // HACK: -90 flips default character model

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
	// Draw character and equipments
	rl.PushMatrix()
	{
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
		characterRotate = rl.QuaternionFromAxisAngle(rl.NewVector3(0.0, 1.0, 0.0), float32(characterAngle)*rl.Deg2rad)
		characterModel.Transform = rl.MatrixMultiply(rl.QuaternionToMatrix(characterRotate), rl.MatrixTranslate(posX, posY, posZ))
		rl.UpdateModelAnimation(characterModel, anim, int32(animCurrentFrame))
		rl.DrawMesh(characterModel.GetMeshes()[0],
			characterModel.GetMaterials()[1], characterModel.Transform)

		// Draw equipments (hat, sword, shield)
		for i := range MaxBoneSockets {
			if !isShowEquippedModels[i] {
				continue
			}
			if anim.FramePoses == nil || characterModel.BindPose == nil {
				continue
			}

			var transform rl.Transform = anim.GetFramePose(int(animCurrentFrame), boneSocketsIndex[i])
			var inRotation rl.Quaternion = characterModel.GetBindPose()[boneSocketsIndex[i]].Rotation
			var outRotation rl.Quaternion = transform.Rotation

			// Calculate socket rotation (angle between bone in initial pose and same bone in current animation frame)
			var rotate rl.Quaternion = rl.QuaternionMultiply(outRotation, rl.QuaternionInvert(inRotation))
			var matrixTransform rl.Matrix = rl.QuaternionToMatrix(rotate)
			// Translate socket to its position in the current animation
			matrixTransform = rl.MatrixMultiply(matrixTransform, rl.MatrixTranslate(transform.Translation.X, transform.Translation.Y, transform.Translation.Z))
			// Transform the socket using the transform of the character (angle and translate)
			matrixTransform = rl.MatrixMultiply(matrixTransform, characterModel.Transform)

			// Draw mesh at socket position with socket angle rotation
			rl.DrawMesh(equippedModels[i].GetMeshes()[0], equippedModels[i].GetMaterials()[1], matrixTransform)
		}
	}
	rl.PopMatrix()

	// Debug
	if false {
		size := rl.Vector3Scale(p.Size, .5)

		// Debug player color near player's feet
		rl.DrawCubeV(rl.Vector3Subtract(p.Position, rl.NewVector3(0., p.Size.Y, 0.)), p.Size, rl.Fade(playerColor, .3))

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

func ToggleEquippedModels(values [MaxBoneSockets]bool) {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	for i := range values {
		isShowEquippedModels[i] = values[i]
	}
}

// FIXME: Camera gets stuck if player keeps moving into the box.
// NOTE:  Maybe lerp or free camera if "distance to the box is less" or touching.
// PERF: https://github.com/raylib-extras/examples-c/blob/6ed2ac244d961239b1695d0b6a729f6fd7bc209b/platformer_motion/platformer.c#L34C1-L147C2
//
//	checks a moving rectangle against some static object, stopping them motion based on what side of the static object is hit.
//	hit booleans return back what part of the object was hit to help with state collisons
//	void CollideRectWithObject(const Rectangle mover, const Rectangle object, Vector2* motion, bool* hitSide, bool* hitTop, bool* hitBottom)
func RevertPlayerAndCameraPositions(
	dstPlayer *Player,
	srcPlayer Player,
	dstCamera *rl.Camera3D,
	srcCamera rl.Camera3D,
) {
	dstPlayer.Position = srcPlayer.Position
	dstPlayer.BoundingBox = common.GetBoundingBoxFromPositionSizeV(dstPlayer.Position, dstPlayer.Size)
	dstCamera.Target = srcCamera.Target
	dstCamera.Position = srcCamera.Position
}
