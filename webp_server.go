package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/gofiber/fiber"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Config struct {
	HOST          string
	PORT          string
	IMG_PATH      string
	QUALITY       string
	ALLOWED_TYPES []string
}

func webpEncoder(p1, p2 string, quality float32) {
	var buf bytes.Buffer
	var img image.Image
	data, _ := ioutil.ReadFile(p1)
	if strings.Contains(p1, "jpg") || strings.Contains(p1, "jpeg") {
		img, _ = jpeg.Decode(bytes.NewReader(data))
	} else if strings.Contains(p1, "png") {
		img, _ = png.Decode(bytes.NewReader(data))
	}

	if err := webp.Encode(&buf, img, &webp.Options{Lossless: true, Quality: quality}); err != nil {
		log.Println(err)
	}
	if err := ioutil.WriteFile(p2, buf.Bytes(), 0666); err != nil {
		log.Println(err)
	}

	fmt.Println("Save output.webp ok")
}

func main() {
	app := fiber.New()
	app.Banner = false
	app.Server = "WebP Server Go"

	// Config Here
	config := load_config("config.json")

	HOST := config.HOST
	PORT := config.PORT
	IMG_PATH := config.IMG_PATH
	QUALITY := config.QUALITY
	ALLOWED_TYPES := config.ALLOWED_TYPES

	LISTEN_ADDRESS := HOST + ":" + PORT

	// Server Info
	SERVER_INFO := "WebP Server is running at " + LISTEN_ADDRESS
	fmt.Println(SERVER_INFO)

	app.Get("/*", func(c *fiber.Ctx) {

		// /var/www/IMG_PATH/path/to/tsuki.jpg
		IMG_ABSOLUTE_PATH := IMG_PATH + c.Path()

		// /path/to/tsuki.jpg
		IMG_PATH := c.Path()

		// jpg
		IMG_EXT := strings.Split(path.Ext(IMG_PATH), ".")[1]

		// tsuki.jpg
		IMG_NAME := path.Base(IMG_PATH)

		// /path/to
		DIR_PATH := path.Dir(IMG_PATH)

		// /path/to/tsuki.jpg.webp
		WEBP_IMG_PATH := DIR_PATH + "/" + IMG_NAME + ".webp"

		// /home/webp_server
		CURRENT_PATH, err := os.Getwd()
		if err != nil {
			fmt.Println(err.Error())
		}

		// /home/webp_server/exhaust/path/to/tsuki.webp
		WEBP_ABSOLUTE_PATH := CURRENT_PATH + "/exhaust" + WEBP_IMG_PATH

		// /home/webp_server/exhaust/path/to
		DIR_ABSOLUTE_PATH := CURRENT_PATH + "/exhaust" + DIR_PATH

		// Check file extension
		_, found := Find(ALLOWED_TYPES, IMG_EXT)
		if !found {
			c.Send("File extension not allowed!")
			c.SendStatus(403)
			return
		}

		// Check the original image for existence
		if !imageExists(IMG_ABSOLUTE_PATH) {
			// The original image doesn't exist, check the webp image, delete if processed.
			if imageExists(WEBP_ABSOLUTE_PATH) {
				os.Remove(WEBP_ABSOLUTE_PATH)
			}
			c.Send("File not found!")
			c.SendStatus(404)
			return
		}

		if imageExists(WEBP_ABSOLUTE_PATH) {
			c.SendFile(WEBP_ABSOLUTE_PATH)
		} else {
			// Mkdir
			os.MkdirAll(DIR_ABSOLUTE_PATH, os.ModePerm)

			// cwebp -q 60 Cute-Baby-Girl.png -o Cute-Baby-Girl.webp

			q, _ := strconv.ParseFloat(QUALITY, 32)
			webpEncoder(IMG_ABSOLUTE_PATH, WEBP_ABSOLUTE_PATH, float32(q))
			if err != nil {
				fmt.Println(err)
			}
			c.SendFile(WEBP_ABSOLUTE_PATH)
		}
	})

	app.Listen(LISTEN_ADDRESS)
}

func load_config(path string) Config {
	var config Config
	json_object, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer json_object.Close()
	decoder := json.NewDecoder(json_object)
	decoder.Decode(&config)
	return config
}

func imageExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}