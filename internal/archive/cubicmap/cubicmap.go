package cubicmap

import (
	"example/depths/internal/common"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawCubicmaps() {
	// Draw cube with an applied texture
	vec := rl.Vector3{X: -2.0, Y: 2.0}
	DrawCubeTexture(common.Texture.CubicmapAtlas, vec, 2.0, 4.0, 2.0, rl.White)

	// Draw cube with an applied texture, but only a defined rectangle piece of the texture
	rec := rl.Rectangle{
		Y:      float32(common.Texture.CubicmapAtlas.Height) / 2.0,
		Width:  float32(common.Texture.CubicmapAtlas.Width) / 2.0,
		Height: float32(common.Texture.CubicmapAtlas.Height) / 2.0,
	}
	vec = rl.Vector3{X: 2.0, Y: 1.0}
	DrawCubeTextureRec(common.Texture.CubicmapAtlas, rec, vec, 2.0, 2.0, 2.0, rl.White)
}

// DrawCubeTexture draws a textured cube
// NOTE: Cube position is the center position
func DrawCubeTexture(texture rl.Texture2D, position rl.Vector3, width, height, length float32, color rl.Color) {
	const isEnableCommentedCode = false

	x := position.X
	y := position.Y
	z := position.Z

	rl.SetTexture(texture.ID)
	if isEnableCommentedCode { // Set desired texture to be enabled while drawing following vertex data
		// Vertex data transformation can be defined with the commented lines,
		// but in this example we calculate the transformed vertex data directly when calling rlVertex3f()
		rl.PushMatrix()
		// NOTE: Transformation is applied in inverse order (scale -> rotate -> translate)
		rl.Translatef(2.0, 0.0, 0.0)
		rl.Rotatef(45, 0, 1, 0)
		rl.Scalef(2.0, 2.0, 2.0)
	}

	rl.Begin(rl.Quads)
	rl.Color4ub(color.R, color.G, color.B, color.A)
	// Front Face
	rl.Normal3f(0.0, 0.0, 1.0) // Normal Pointing Towards Viewer
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	// Back Face
	rl.Normal3f(0.0, 0.0, -1.0) // Normal Pointing Away From Viewer
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	// Top Face
	rl.Normal3f(0.0, 1.0, 0.0) // Normal Pointing Up
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad.
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad Bottom Face
	rl.Normal3f(0.0, -1.0, 0.0)                    // Normal Pointing Down
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	// Right face
	rl.Normal3f(1.0, 0.0, 0.0) // Normal Pointing Right
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2) // Top Left Of The Texture and Quad
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2) // Bottom Left Of The Texture and Quad
	// Left Face
	rl.Normal3f(-1.0, 0.0, 0.0) // Normal Pointing Left
	rl.TexCoord2f(0.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2) // Bottom Left Of The Texture and Quad
	rl.TexCoord2f(1.0, 0.0)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2) // Bottom Right Of The Texture and Quad
	rl.TexCoord2f(1.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2) // Top Right Of The Texture and Quad
	rl.TexCoord2f(0.0, 1.0)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2) // Top Left Of The Texture and Quad

	rl.End()

	if isEnableCommentedCode {
		rl.PopMatrix()
	}

	rl.SetTexture(0)
}

// DrawCubeTextureRec draws a cube with texture piece applied to all faces
func DrawCubeTextureRec(
	texture rl.Texture2D,
	source rl.Rectangle,
	position rl.Vector3,
	width, height,
	length float32,
	color rl.Color,
) {
	x := position.X
	y := position.Y
	z := position.Z

	texWidth := float32(texture.Width)
	texHeight := float32(texture.Height)

	// Set desired texture to be enabled while drawing following vertex data
	rl.SetTexture(texture.ID)

	// We calculate the normalized texture coordinates for the desired texture-source-rectangle
	// It means converting from (tex.width, tex.height) coordinates to [0.0f, 1.0f] equivalent
	rl.Begin(rl.Quads)
	rl.Color4ub(color.R, color.G, color.B, color.A)

	// Front face
	rl.Normal3f(0.0, 0.0, 1.0)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)

	// Back face
	rl.Normal3f(0.0, 0.0, -1.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)

	// Top face
	rl.Normal3f(0.0, 1.0, 0.0)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)

	// Bottom face
	rl.Normal3f(0.0, -1.0, 0.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)

	// Right face
	rl.Normal3f(1.0, 0.0, 0.0)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z-length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x+width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x+width/2, y-height/2, z+length/2)

	// Left face
	rl.Normal3f(-1.0, 0.0, 0.0)
	rl.TexCoord2f(source.X/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z-length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, (source.Y+source.Height)/texHeight)
	rl.Vertex3f(x-width/2, y-height/2, z+length/2)
	rl.TexCoord2f((source.X+source.Width)/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z+length/2)
	rl.TexCoord2f(source.X/texWidth, source.Y/texHeight)
	rl.Vertex3f(x-width/2, y+height/2, z-length/2)

	rl.End()

	rl.SetTexture(0)
}
