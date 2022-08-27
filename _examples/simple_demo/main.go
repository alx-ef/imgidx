package main

import (
	"github.com/alef-ru/imgidx"
	"github.com/alef-ru/imgidx/embedders"
	"log"
	"path/filepath"
	"strings"
)

type ImgAttrs struct {
	Name string
	Type string
}

var imagePaths = []string{
	"testdata/pokemon/abomasnow.png",
	"testdata/pokemon/abra.png",
	"testdata/pokemon/absol.png",
	"testdata/pokemon/accelgor.png",
	"testdata/pokemon/aegislash-blade.png",
	"testdata/pokemon/aerodactyl.png",
	"testdata/pokemon/aggron.png",
	"testdata/pokemon/aipom.png",
	"testdata/pokemon/alakazam.png",
	"testdata/pokemon/alomomola.png",
	"testdata/pokemon/altaria.png",
	"testdata/pokemon/amaura.png",
	"testdata/pokemon/ambipom.png",
	"testdata/pokemon/amoonguss.png",
	"testdata/pokemon/ampharos.png",
	"testdata/pokemon/anorith.png",
	"testdata/pokemon/araquanid.jpg",
	"testdata/pokemon/arbok.png",
	"testdata/pokemon/arcanine.png",
	"testdata/pokemon/arceus.png",
}

func main() {
	// Create non-persistent index with default embedder
	idx, err := imgidx.NewCompositeIndex(8, 8)
	if err != nil {
		log.Fatal(err)
	}

	// Add images to index
	for _, path := range imagePaths {
		name := filepath.Base(path)
		_, err := imgidx.AddImageFile(idx, path, ImgAttrs{name, "Pokemon"})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s added to the index\n", path)
	}

	// Search for similar images in the index
	log.Printf("Exactly the same file:")
	printNearestImage(idx, "testdata/pokemon/abomasnow.png")

	log.Printf("Convrted PNG to JPEG and compressed image:")
	printNearestImage(idx, "testdata/compressed_abomasnow.jpg")

	log.Printf("Severly altered version of the image: " +
		"It has different size and proportions, different colors, some extra curves and frame")
	printNearestImage(idx, "testdata/distorted_abomasnow.jpg")

	log.Printf("Now we delete abomasnow.png from the index and try to find the nearest image to it")
	idx.Remove(func(vec embedders.Vector, uri string, attrs interface{}) bool {
		return strings.HasSuffix(uri, "/abomasnow.png")
	})
	printNearestImage(idx, "testdata/pokemon/abomasnow.png")
}

func printNearestImage(idx imgidx.Index, path string) {
	_, attrs, dist, err := imgidx.NearestByFile(idx, path)
	if err != nil {
		log.Fatal(err)
	}
	imgAttrs := attrs.(ImgAttrs)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Nearest image to %s : %s (distacne: %f)\n", path, imgAttrs.Name, dist)
}
