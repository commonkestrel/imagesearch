//A package designed to search Google Images based on the input query and arguments. These images may be protected under copyright, and you shouldn't do anything punishable with them, like using them for commercial use. This library is inspired by the Python library [google-images-download] created by [hardikvasa] and maintained by [joeclinton1], but ported to Go and with some quality of life improvements, such as being able to retrieve urls as well. Essentially, this package is a port of the Python library [GoogleImageScraper] to Go.
//
//[google-images-download]: https://www.github.com/joeclinton1/google-images-download
//[hardikvasa]: https://www.github.com/hardikvasa
//[joeclinton1]: https://www.github.com/joeclinton1
//[GoogleImageScraper]: https://www.github.com/Jibble330/GoogleImageScraper
package imagesearch

import (
    "os"
    "io"
    "path"
    "html"
    "errors"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
    "path/filepath"
)

//Contains information about an image including the url of the image, the url of the source, and the website it came from. Example: 
//  Image {
//      Url: www.example.com/static/image.png
//      Source: www.example.com/article
//      Base: example.com
//  }
//
type Image struct {
    Url       string `json:"url"`
    Source    string `json:"source"`
    Base      string `json:"base"`
}

//These variables are all of the possible arguments that can be passed into Images, Download, and Urls. These are used by passing gis.{Argument}.{Option} into the arguments parameter. For Example: 
//  urls, err := gis.Urls("example", 0, gis.Color.Red, gis.License.CreativeCommons)
var (
    Color = struct {
        Red, Orange, Yellow, Green, Teal, Blue, Purple, Pink, White, Gray, Black, Brown string
    }{Red: "isc:red", Orange: "isc:orange", Yellow: "isc:yellow", Green: "isc:green", Teal: "isc:teel", Blue: "isc:blue", Purple: "isc:purple", Pink: "isc:pink", White: "isc:white", Gray: "isc:gray", Black: "isc:black", Brown: "isc:brown"}

    ColorType = struct{
        Color, Grayscale, Transparent string
    }{Color: "ic:full", Grayscale: "ic:gray", Transparent: "ic:trans"}

    License = struct{
        CreativeCommons, Other string
    }{CreativeCommons: "il:cl", Other: "il:ol"}

    Type = struct{
        Face, Photo, Clipart, Lineart, Animated string
    }{Face: "itp:face", Photo: "itp:photo", Clipart: "itp:clipart", Lineart: "itp:lineart", Animated: "itp:animated"}

    Time = struct {
        PastDay, PastWeek, PastMonth, PastYear string
    }{PastDay: "qdr:d", PastWeek: "qdr:w", PastMonth: "qdr:m", PastYear:"qdr:y"}

    AspectRatio = struct{
        Tall, Square, Wide, Panoramic string
    }{Tall: "iar:t", Square: "iar:s", Wide: "iar:w", Panoramic: "iar:xw"}

    Format = struct{
        Jpg, Gif, Png, Bmp, Svg, Webp, Ico, Raw string
    }{Jpg: "ift:jpg", Gif: "ift:gif", Png: "ift:png", Bmp: "ift:bmp", Svg: "ift:svg", Webp: "webp", Ico: "ift:ico", Raw: "ift:craw"}
)


//Searches for the query along with the given arguments, and returns a slice of Image objects.
//The amount of images does not exceed the limit unless the limit is 0, in which case it will return all images found.
func Images(query string, limit int, arguments ...string) ([]Image, error) {
    url := buildUrl(query, arguments)

    page, err := getPage(url)
    if err != nil {
        return []Image{}, err
    }

    images, err := findImages(page)
    if err != nil {
        return []Image{}, err
    }

    if len(images) > limit && limit > 0 {
        images = images[:limit]
    }

    return images, nil
}


//Searches for the query along with the given arguments, and returns a slice of the image urls.
//The amount of images does not exceed the limit unless the limit is 0, in which case it will return all urls found.
func Urls(query string, limit int, arguments ...string) ([]string, error) {
    url := buildUrl(query, arguments)

    page, err := getPage(url)
    if err != nil {
        return []string{}, err
    }

    images, err := findImages(page)
    if err != nil {
        return []string{}, err
    }

    if len(images) > limit && limit > 0 {
        images = images[:limit]
    }

    var urls []string
    for _, image := range images {
        urls = append(urls, image.Url)
    }

    return urls, nil
}

//Searches for the given query along with the given argumetnts and downloads the images into the given directory.
//The amount of images does not exceed the limit unless the limit is 0, in which case it will download all images found.
//Returns a slice of the absolute paths of all downloaded images.
func Download(query string, limit int, dir string, arguments ...string) ([]string, int, error) {
    dir = strings.ReplaceAll(dir, "\\", "/")
    urls, err := Urls(query, limit, arguments...)
    if err != nil {
        return []string{}, 0, nil
    }

    var suffix int
    var errs int
    var paths []string

    for _, url := range urls {
        pat := path.Join(dir, query+strconv.Itoa(suffix)) + ".*"
        matches, _ := filepath.Glob(pat)
        for len(matches) > 0 {
            suffix++
            pat = path.Join(dir, query+strconv.Itoa(suffix)) + ".*"
            matches, _ = filepath.Glob(pat)
        }

        file, err := DownloadImage(url, dir, query+strconv.Itoa(suffix))
        if err != nil {
            errs++
        }

        paths = append(paths, file)
    }

    return paths, errs, nil
}

//Given the url of the image, the directory to download to, and the name of the file *without extension*, this will find the type of image and download it to the given directory.
//Warning: This will overwrite any image file with the same name, if the extension matches, so make sure to keep the name unique.
//You can check if a file with the name already exists with the following code: 
//
//  import (
//      "path"
//      "path/filepath"
//  )
//  
//  func exists(dir, name string) bool {
//      pat := path.Join(dir, name) + ".*"
//      matches, _ := filepath.Glob(pat)
//      return len(matches) > 0
//  }
func DownloadImage(url, dir, name string) (string, error) {
    client := http.DefaultClient
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }

    bytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    mimetype := http.DetectContentType(bytes)
    var extension string
    if strings.Contains(mimetype, "image") {
        extension = strings.ReplaceAll(mimetype, "image/", "")
    } else {
        return "", errors.New("invalid image format")
    }

    file := name + "." + extension
    abs := path.Join(dir, file)

    f, err := os.Create(abs)
    if err != nil {
        return "", err
    }
    _, err = f.Write(bytes)
    if err != nil {
        return "", err
    }

    return f.Name(), nil
}

func buildUrl(query string, arguments []string) string {
    url := "https://www.google.com/search?tbm=isch&q=" + query

    if len(arguments) > 0 {
        url += "ic:specific"
    }
    for _, argument := range arguments {
        url += "%2C" + argument
    }

    return url
}

func findImages(page string) ([]Image, error) {
    scriptStart := strings.LastIndex(page, "AF_initDataCallback")
    page = page[scriptStart:]

    startChar := strings.Index(page, "[")
    page= page[startChar:]

    endChar := strings.Index(page, "</script>") - 20
    page = page[:endChar]

    var imageJson []interface{}
    
    err := json.Unmarshal([]byte(html.UnescapeString(page)), &imageJson)
    if err != nil {
        return []Image{}, err
    }

    imageObjects := imageJson[56].([]interface{})[1].([]interface{})[0].([]interface{})[0].([]interface{})[1].([]interface{})[0].([]interface{})

    var images []Image
    for _, imageObject := range imageObjects {
        obj := imageObject.([]interface{})[0].([]interface{})[0].(map[string]interface{})["444383007"].([]interface{})[1]
        if obj != nil {
            var image Image
            image.Url = obj.([]interface{})[3].([]interface{})[0].(string)

            sourceInfo := obj.([]interface{})[9].(map[string]interface{})["2003"].([]interface{})
            image.Source = sourceInfo[2].(string)
            image.Base = sourceInfo[17].(string)
            images = append(images, image)
        }
    }
    return images, nil
}

func getPage(url string) (string, error) {
    client := http.DefaultClient
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    html, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    return string(html), nil
}

