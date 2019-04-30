package grm

import (
	"bufio"
	"image"
	"image/draw"
	_ "image/png"
	"io/ioutil"
	"os"
	"unsafe"
)

func readFile(path string) []byte {
	dat, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	return dat
}

func readFileLines(path string) []string {
	var dat []string

	file, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	scr := bufio.NewScanner(file)

	for scr.Scan() {
		dat = append(dat, scr.Text())
	}

	return dat
}

func readFileString(path string) string {
	var dat string

	file, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	scr := bufio.NewScanner(file)

	for scr.Scan() {
		dat += scr.Text() + "\n"
	}

	return dat
}

func getFloatSize(dat []float32) int {
	return int(unsafe.Sizeof(dat[0])) * len(dat)
}

func getIntSize(dat []uint16) int {
	return int(unsafe.Sizeof(dat[0])) * len(dat)
}

type sliceHeader struct {
	data uintptr
	len  int
	cap  int
}

func sliceUint32(dat []byte) []uint32 {
	const m = 0x7fffffff
	return (*[m / 4]uint32)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&dat)).data))[:len(dat)/4]
}

func floatData(dat []float32) []byte {
	if len(dat) == 0 {
		return nil
	}
	const m = 0x7fffffff
	s := int(unsafe.Sizeof(dat[0]))
	return (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&dat)).data))[:len(dat)*s]
}

func intData(dat []uint16) []byte {
	if len(dat) == 0 {
		return nil
	}
	const m = 0x7fffffff
	s := int(unsafe.Sizeof(dat[0]))
	return (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&dat)).data))[:len(dat)*s]
}

func uboData(dat UniformBufferObject) []byte {
	const mm = 0x7fffffff
	s := int(unsafe.Sizeof(dat))
	return (*[mm]byte)(unsafe.Pointer(&dat))[:s]
}

func unsafeByte(us unsafe.Pointer, s int) []byte {
	const mm = 0x7fffffff
	return (*[mm]byte)(unsafe.Pointer(&us))[:s]
}

func loadImageFile(path string) (image.Image, error) {
	infile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer infile.Close()

	img, _, err := image.Decode(infile)
	return img, err
}

func loadTextureData(path string, rowPitch int) ([]byte, uint32, uint32, error) {
	img, err := loadImageFile(path)
	if err != nil {
		return nil, 0, 0, err
	}
	newImg := image.NewRGBA(img.Bounds())
	if rowPitch <= 4*img.Bounds().Dy() {
		newImg.Stride = rowPitch
	}
	draw.Draw(newImg, newImg.Bounds(), img, image.ZP, draw.Src)
	size := newImg.Bounds().Size()
	return []byte(newImg.Pix), uint32(size.X), uint32(size.Y), nil
}
