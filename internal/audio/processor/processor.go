// See https://github.com/gen2brain/raylib-go/blob/master/examples/audio/mixed_processor/main.go
//
//	func Init() {
//		rl.AttachAudioMixedProcessor(audiopro.ProcessAudio)
//		audiopro.InitAudioProcessor()
//	}
//
//	func DeInit() {
//		rl.DetachAudioMixedProcessor(ProcessAudio) // Disconnect audio processor
//	}
package processor

import (
	"example/depths/internal/util/mathutil"
)

var (
	AudioExponent  float32      = 0.5 // Audio exponentiation value [0..1]
	AudioAvgVolume [400]float32       // Average volume history
)

func InitAudioProcessor() {
	AudioExponent = 1.0
	for i := range AudioAvgVolume {
		AudioAvgVolume[i] = 0
	}
}

// ProcessAudio is the audio processing function.
func ProcessAudio(buffer []float32, frames int) {
	var avg float32 // Temporary average volume
	maxFrame := frames / 2

	// Each frame has 2 samples (left and right),
	// so we should loop `frames / 2` times
	for frame := 0; frame < maxFrame; frame++ {
		left := &buffer[frame*2+0]  // Left channel
		right := &buffer[frame*2+1] // Right channel

		// Modify left and right channel samples with exponent
		*left = mathutil.PowF(mathutil.AbsF(*left), AudioExponent) * mathutil.SignF(*left)
		*right = mathutil.PowF(mathutil.AbsF(*right), AudioExponent) * mathutil.SignF(*right)

		// Accumulate average volume
		avg += mathutil.AbsF(*left) / float32(maxFrame)
		avg += mathutil.AbsF(*right) / float32(maxFrame)
	}

	// Shift average volume history buffer to the left
	for i := 0; i < 399; i++ {
		AudioAvgVolume[i] = AudioAvgVolume[i+1]
	}

	// Add the new average value
	AudioAvgVolume[399] = avg
}
