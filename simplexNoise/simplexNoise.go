package main

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

var winWidht = 800
var winHeight = 600

const renderScale = 0.01

func (para Object) drawBlock(pixels []byte) {
	para.x *= float32(winHeight) * renderScale
	para.y *= float32(winHeight) * renderScale
	para.halfLen *= float32(winHeight) * renderScale
	para.halfBrth *= float32(winHeight) * renderScale

	para.x += float32(winWidht) * 0.5
	para.y += float32(winHeight) * 0.5

	x0 := int(para.x - para.halfLen)
	x1 := int(para.x + para.halfLen)
	y0 := int(para.y - para.halfBrth)
	y1 := int(para.y + para.halfBrth)
	drawBlockInPixels(x0, y0, x1, y1, 0, para.color, pixels)
}
func drawBlockInPixels(x0, y0, x1, y1, borderInPix int, color color, pixels []byte) {
	x0 = clampInt(x0, 0, winWidht)
	y0 = clampInt(y0, 0, winHeight)
	x1 = clampInt(x1, x0, winWidht)
	y1 = clampInt(y1, y0, winHeight)

	index := y0 * winWidht
	if borderInPix == 0 {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				pixels[(index+x)*4] = color.r
				pixels[(index+x)*4+1] = color.g
				pixels[(index+x)*4+2] = color.b
			}
			index += winWidht
		}
	} else {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if y < (y0+borderInPix) || y > (y1-borderInPix) || x < (x0+borderInPix) || x > (x1-borderInPix) {
					pixels[index] = 0
					pixels[index+1] = 0
					pixels[index+2] = 0
					continue
				}
				pixels[index] = color.r
				pixels[index+1] = color.g
				pixels[index+2] = color.b
			}
			index += winWidht
		}
	}
}

//func (para Object)noiseBlock(pixels []byte) {
//	para.x *= float32(winHeight) * renderScale
//	para.y *= float32(winHeight) * renderScale
//	para.halfLen *= float32(winHeight) * renderScale
//	para.halfBrth *= float32(winHeight) * renderScale
//
//	para.x += float32(winWidht) * 0.5
//	para.y += float32(winHeight) * 0.5
//
//	x0 := int(para.x - para.halfLen)
//	x1 := int(para.x + para.halfLen)
//	y0 := int(para.y - para.halfBrth)
//	y1 := int(para.y + para.halfBrth)
//	noiseBlockInPixels(x0, y0, x1, y1, para.color, pixels)
//}
//func noiseBlockInPixels(x0, y0, x1, y1 int, color color, pixels []byte)  {
//	x0 = clampInt(x0, 0, winWidht)
//	y0 = clampInt(y0, 0, winHeight);
//	x1 = clampInt(x1, x0, winWidht);
//	y1 = clampInt(y1, y0, winHeight);
//
//	index := y0 * winWidht
//
//	for y := y0; y < y1 ; y++{
//		for x := x0; x < x1; x++{
//			pixels[(index + x)*4] = color.r
//			pixels[(index + x)*4 + 1] = color.g
//			pixels[(index + x)*4 + 2] = color.b
//		}
//		index += winWidht;
//	}
//}
func reScaleAndDraw(noise []float32, min, max float32, gradient []color, pixels []byte) {
	scale := 255.0 / (max - min)
	offset := min * scale
	var p int
	var c color
	for i := range noise {
		noise[i] = noise[i]*scale - offset
		c = gradient[(clampInt(int(noise[i]), 0, 255))]
		p = i * 4
		pixels[p] = c.r
		pixels[p+1] = c.g
		pixels[p+2] = c.b
	}
}
func turbulance(x, y, freq, lac, gain float32, octaves int) float32 {
	var sum, f float32
	ampltude := float32(1.0)
	for i := 0; i < octaves; i++ {
		f = snoise2(x*freq, y*freq) * ampltude
		if f < 0 {
			f = -f
		}
		sum += f
		freq *= lac
		ampltude *= gain
	}
	return sum
}
func fbm2(x, y, freq, lac, gain float32, octaves int) float32 {
	var sum float32
	ampltude := float32(1.0)
	for i := 0; i < octaves; i++ {
		sum += snoise2(x*freq, y*freq) * ampltude
		freq *= lac
		ampltude *= gain
	}
	return sum
}
func makeNoise(pixels []byte, freq, lac, gain float32, octaves, w, h int) {
	var mutex = &sync.Mutex{}
	noise := make([]float32, winHeight*winWidht)

	var min float32
	var max float32

	numRoute := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numRoute)
	batchSize := len(noise) / numRoute

	for i := 0; i < numRoute; i++ {
		go func(i int) {
			defer wg.Done()
			start := i * batchSize
			end := start + batchSize - 1
			var x, y int
			for j := start; j < end; j++ {
				x = j % w
				y = (j - x) / w
				noise[j] = turbulance(float32(x), float32(y), freq, lac, gain, octaves)
				//since we are locking only when we have to change min & max so why should we lock other times as lock and unlock takes time(it lock entire execution of every routine)
				mutex.Lock() //here we lock in order to allow only one routine to go at a time for safe routine
				if noise[j] < min {
					min = noise[j]
				} else if noise[j] > max {
					max = noise[j]
				}
				mutex.Unlock() //here we unlock
			}
		}(i) //(here for min & max but we are not using it as input beacause every routine needs to see max & min rather here we would lock unlock)for safe routine means if any value that is outside of routine is getting changed then there is possibility of affecting other routine since function is only one it is just executing simultaneously by many threads so if any outside value is changed while priveous routine is still woking then that will generate (invisible)error in data
	}
	//makeNoise is in main thread while routines are in other(and main) threads so it is possible for routines to take some time and if routine is taking time then itis necessary to wait routines to finish what they are doing especially when they are making impact(providing data) for further work
	wg.Wait()
	gradient := getDualGradient(color{0, 0, 255}, color{0, 255, 0}, color{255, 255, 0}, color{255, 0, 0})
	reScaleAndDraw(noise, min, max, gradient, pixels)
}
func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}
func colorLerp(c1, c2 color, pct float32) color {
	return color{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}
func getGradient(c1, c2 color) []color {
	result := make([]color, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}
func getDualGradient(c1, c2, c3, c4 color) []color {
	result := make([]color, 256)
	var pct float32
	for i := range result {
		pct = float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*2)
		} else {
			pct -= 0.5
			result[i] = colorLerp(c3, c4, pct*2)
		}
	}
	return result
}
func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

type color struct {
	r, g, b byte
}

//Object is parametres for drawing
type Object struct {
	x, y              float32
	halfLen, halfBrth float32
	color             color
}
type element struct {
	Object
	speedX, speedY float32
	accln          int8
}

func clearScreen(color color, pixels []byte) {
	var index int = 0
	for y := 0; y < winHeight; y++ {
		for x := 0; x < winWidht; x++ {
			pixels[(index+x)*4] = color.r
			pixels[(index+x)*4+1] = color.g
			pixels[(index+x)*4+2] = color.b
		}
		index += winWidht
	}
}

func main() {
	window, err := sdl.CreateWindow("Test sdl2", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, int32(winWidht), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy() //defer executes this statement after reaching the end of function/finishing the execution of funtion

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(winWidht), int32(winHeight))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tex.Destroy()

	pixels := make([]byte, winWidht*winHeight*4)

	keyState := sdl.GetKeyboardState()
	var freq, lac, gain float32
	var octaves int
	var fact int8
	freq, lac, gain, octaves, fact = .0005, 2.5, 0.7, 9, 1
	makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		//simulate
		//keyState = sdl.GetKeyboardState()
		//Object{0,0,40,50,color{200,150,0}}.noiseBlock(pixels)
		if keyState[sdl.SCANCODE_LSHIFT] != 0 || keyState[sdl.SCANCODE_RSHIFT] != 0 {
			fact = 5
		} else {
			fact = 1
		}
		if keyState[sdl.SCANCODE_F] != 0 && keyState[sdl.SCANCODE_DOWN] != 0 {
			freq -= 0.0001 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		} else if keyState[sdl.SCANCODE_F] != 0 && keyState[sdl.SCANCODE_UP] != 0 {
			freq += 0.0001 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		}
		if keyState[sdl.SCANCODE_L] != 0 && keyState[sdl.SCANCODE_DOWN] != 0 {
			lac -= 0.1 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		} else if keyState[sdl.SCANCODE_L] != 0 && keyState[sdl.SCANCODE_UP] != 0 {
			lac += 0.1 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		}
		if keyState[sdl.SCANCODE_G] != 0 && keyState[sdl.SCANCODE_DOWN] != 0 {
			gain -= 0.1 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		} else if keyState[sdl.SCANCODE_G] != 0 && keyState[sdl.SCANCODE_UP] != 0 {
			gain += 0.1 * float32(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		}
		if keyState[sdl.SCANCODE_O] != 0 && keyState[sdl.SCANCODE_DOWN] != 0 {
			octaves -= int(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		} else if keyState[sdl.SCANCODE_O] != 0 && keyState[sdl.SCANCODE_UP] != 0 {
			octaves += int(fact)
			makeNoise(pixels, freq, lac, gain, octaves, winWidht, winHeight)
			fmt.Println("freq:", freq, "lacunarity:", lac, "gain:", gain, "octaves:", octaves)
		}
		//render
		tex.Update(nil, pixels, winWidht*4)
		renderer.Copy(tex, nil, nil)
		renderer.Present()
		sdl.Delay(100)
	}

}

// NOISE FUNCTIONS

func fastFloor(x float32) int {
	if float32(int(x)) <= x {
		return int(x)
	}
	return int(x) - 1
}

// Static data

/*
 * Permutation table. This is just a random jumble of all numbers 0-255
 * This needs to be exactly the same for all instances on all platforms,
 * so it's easiest to just keep it as static explicit data.
 * This also removes the need for any initialisation of this class.
 *
 */
var perm = [256]uint8{151, 160, 137, 91, 90, 15,
	131, 13, 201, 95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23,
	190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33,
	88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139, 48, 27, 166,
	77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244,
	102, 143, 54, 65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200, 196,
	135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64, 52, 217, 226, 250, 124, 123,
	5, 202, 38, 147, 118, 126, 255, 82, 85, 212, 207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42,
	223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9,
	129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228,
	251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241, 81, 51, 145, 235, 249, 14, 239, 107,
	49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254,
	138, 236, 205, 93, 222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180}

//---------------------------------------------------------------------

func grad2(hash uint8, x, y float32) float32 {
	h := hash & 7 // Convert low 3 bits of hash code
	u := y
	v := 2 * x
	if h < 4 {
		u = x
		v = 2 * y
	} // into 8 simple gradient directions,
	// and compute the dot product with (x,y).

	if h&1 != 0 {
		u = -u
	}
	if h&2 != 0 {
		v = -v
	}
	return u + v
}

// 2D simplex noise
func snoise2(x, y float32) float32 {

	const F2 float32 = 0.366025403 // F2 = 0.5*(sqrt(3.0)-1.0)
	const G2 float32 = 0.211324865 // G2 = (3.0-Math.sqrt(3.0))/6.0

	var n0, n1, n2 float32 // Noise contributions from the three corners

	// Skew the input space to determine which simplex cell we're in
	s := (x + y) * F2 // Hairy factor for 2D
	xs := x + s
	ys := y + s
	i := fastFloor(xs)
	j := fastFloor(ys)

	t := float32(i+j) * G2
	X0 := float32(i) - t // Unskew the cell origin back to (x,y) space
	Y0 := float32(j) - t
	x0 := x - X0 // The x,y distances from the cell origin
	y0 := y - Y0

	// For the 2D case, the simplex shape is an equilateral triangle.
	// Determine which simplex we are in.
	var i1, j1 uint8 // Offsets for second (middle) corner of simplex in (i,j) coords
	if x0 > y0 {
		i1 = 1
		j1 = 0
	} else { // lower triangle, XY order: (0,0)->(1,0)->(1,1)
		i1 = 0
		j1 = 1
	} // upper triangle, YX order: (0,0)->(0,1)->(1,1)

	// A step of (1,0) in (i,j) means a step of (1-c,-c) in (x,y), and
	// a step of (0,1) in (i,j) means a step of (-c,1-c) in (x,y), where
	// c = (3-sqrt(3))/6

	x1 := x0 - float32(i1) + G2 // Offsets for middle corner in (x,y) unskewed coords
	y1 := y0 - float32(j1) + G2
	x2 := x0 - 1.0 + 2.0*G2 // Offsets for last corner in (x,y) unskewed coords
	y2 := y0 - 1.0 + 2.0*G2

	// Wrap the integer indices at 256, to avoid indexing perm[] out of bounds
	ii := uint8(i)
	jj := uint8(j)

	// Calculate the contribution from the three corners
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 < 0.0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * grad2(perm[ii+perm[jj]], x0, y0)
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 < 0.0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * grad2(perm[ii+i1+perm[jj+j1]], x1, y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 < 0.0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * grad2(perm[ii+1+perm[jj+1]], x2, y2)
	}

	// Add contributions from each corner to get the final noise value.
	return (n0 + n1 + n2)
}
