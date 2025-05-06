package mathutil

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Copied from Go's cmp.Ordered
// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// Note that floating-point types may contain NaN ("not-a-number") values.
// An operator such as == or < will always report false when
// comparing a NaN value with any other value, NaN or not.
// See the [Compare] function for a consistent way to compare NaN values.
type OrderedNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// NumberType typecast to avoid casting `OrderedNumber` interface when used.
type NumberType OrderedNumber

func AbsF[T NumberType](x T) float32       { return float32(math.Abs(float64(x))) }
func SqrtF[T NumberType](x T) float32      { return float32(math.Sqrt(float64(x))) }
func CosF[T NumberType](x T) float32       { return float32(math.Cos(float64(x))) }
func SinF[T NumberType](x T) float32       { return float32(math.Sin(float64(x))) }
func Atan2F[T NumberType](x, y T) float32  { return float32(math.Atan2(float64(x), float64(y))) }
func FloorF[T NumberType](x T) float32     { return float32(math.Floor(float64(x))) }
func CeilF[T NumberType](x T) float32      { return float32(math.Ceil(float64(x))) }
func RoundI[T NumberType](x T) int32       { return int32(math.Round(float64(x))) }
func RoundF[T NumberType](x T) float32     { return float32(math.Round(float64(x))) }
func RoundEvenF[T NumberType](x T) float32 { return float32(math.RoundToEven(float64(x))) }
func SignF[T NumberType](x T) float32 {
	if x == 0 {
		return 0
	}
	return float32(math.Abs(float64(x)) / float64(x))
}
func PingPongF[T NumberType](x T) float32 {
	if x == 0 {
		return 0
	}
	return -SignF(x)
}
func MaxF[T NumberType](x T, y T) float32 { return float32(max(float64(x), float64(y))) }
func MinF[T NumberType](x T, y T) float32 { return float32(min(float64(x), float64(y))) }
func PowF[T NumberType](x T, y T) float32 { return float32(math.Pow(float64(x), float64(y))) }
func MaxI[T NumberType](x T, y T) int32   { return int32(max(float64(x), float64(y))) }
func MinI[T NumberType](x T, y T) int32   { return int32(min(float64(x), float64(y))) }

// Thanks to [hippocoder](https://discussions.unity.com/t/angle-between-camera-and-object/450430/9)
// Best understand the tried and tested old school methods with euler angles. This is the old way everyone has done since time began.
// Behold the code in 2D (which you probably need if you’re not bothered about all the angles):
//
//	function Angle2D(x1:float, y1:float, x2:float, y2:float) {
//		return Mathf.Atan2(y2-y1, x2-x1)*Mathf.Rad2Deg;
//	}
//
// What we do is use Atan2 and plug in two positions, a source and a destination, which is converted to degrees.
// If you want to use a different angle, just plug in x and z for example or y and z. Have fun.
func Angle2D(x1, y1, x2, y2 float32) float32 {
	// const rl.Rad2deg untyped float = 57.295776 // 57.2958
	return Atan2F(y2-y1, x2-x1) * rl.Rad2deg
}
