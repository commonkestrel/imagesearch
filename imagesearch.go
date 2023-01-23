// A package designed to search Google Images based on the input query and arguments. Due to the limitations of using only a single request to fetch images, only a max of about 100 images can be found per request. If you need to find more than 100, one of the many packages using simulated browsers may work better. These images may be protected under copyright, and you shouldn't do anything punishable with them, like using them for commercial use.
package imagesearch

import (
    "encoding/json"
    "errors"
    "html"
    "io"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strconv"
    "strings"
)

var (
    errUnpack = errors.New("failed to unpack json! no image results or Google changed their structrue")
)

// Contains information about an image including the url of the image, the url of the source, and the website it came from. Example:
//
//	Image {
//	    Url: "www.example.com/static/image.png"
//	    Source: "www.example.com/article"
//	    Base: "example.com"
//	}
type Image struct {
    // Image URL
    Url    string `json:"url"`

    // URL the image was found at
    Source string `json:"source"`

    // Base of the source URL
    Base   string `json:"base"`
}

// These variables are all of the possible arguments that can be passed into Images, Download, and Urls. These are used by passing imagesearch.{Argument}.{Option} into the arguments parameter. For example:
//	urls, err := imagesearch.Urls("example", 0, imagesearch.Color.Red, imagesearch.License.CreativeCommons)
var (
    Color = struct {
        Red, Orange, Yellow, Green, Teal, Blue, Purple, Pink, White, Gray, Black, Brown string
    }{Red: "isc:red", Orange: "isc:orange", Yellow: "isc:yellow", Green: "isc:green", Teal: "isc:teel", Blue: "isc:blue", Purple: "isc:purple", Pink: "isc:pink", White: "isc:white", Gray: "isc:gray", Black: "isc:black", Brown: "isc:brown"}

    ColorType = struct {
        Color, Grayscale, Transparent string
    }{Color: "ic:full", Grayscale: "ic:gray", Transparent: "ic:trans"}

    License = struct {
        CreativeCommons, Other string
    }{CreativeCommons: "il:cl", Other: "il:ol"}

    Type = struct {
        Face, Photo, Clipart, Lineart, Animated string
    }{Face: "itp:face", Photo: "itp:photo", Clipart: "itp:clipart", Lineart: "itp:lineart", Animated: "itp:animated"}

    Time = struct {
        PastDay, PastWeek, PastMonth, PastYear string
    }{PastDay: "qdr:d", PastWeek: "qdr:w", PastMonth: "qdr:m", PastYear: "qdr:y"}

    AspectRatio = struct {
        Tall, Square, Wide, Panoramic string
    }{Tall: "iar:t", Square: "iar:s", Wide: "iar:w", Panoramic: "iar:xw"}

    Format = struct {
        Jpg, Gif, Png, Bmp, Svg, Webp, Ico, Raw string
    }{Jpg: "ift:jpg", Gif: "ift:gif", Png: "ift:png", Bmp: "ift:bmp", Svg: "ift:svg", Webp: "webp", Ico: "ift:ico", Raw: "ift:craw"}
)

// Searches for the query along with the given arguments, and returns a slice of Image objects.
// The amount of images does not exceed the limit unless the limit is 0, in which case it will return all images found.
func Images(query string, limit int, arguments ...string) (images []Image, err error) {
    url := buildUrl(query, arguments)

    page, err := getPage(url)
    if err != nil {
        return []Image{}, err
    }

    images, err = unpack(page)
    if err != nil {
        return []Image{}, err
    }

    if len(images) > limit && limit > 0 {
        images = images[:limit]
    }

    return images, nil
}

// Searches for the query along with the given arguments, and returns a slice of the image urls.
// The amount of images does not exceed the limit unless the limit is 0, in which case it will return all urls found.
func Urls(query string, limit int, arguments ...string) (urls []string, err error) {
    url := buildUrl(query, arguments)

    page, err := getPage(url)
    if err != nil {
        return []string{}, err
    }

    images, err := unpack(page)
    if err != nil {
        return []string{}, err
    }

    if len(images) > limit && limit > 0 {
        images = images[:limit]
    }

    for _, image := range images {
        urls = append(urls, image.Url)
    }

    return urls, nil
}

// Searches for the given query along with the given argumetnts and downloads the images into the given directory.
// The amount of images does not exceed the limit unless the limit is 0, in which case it will download all images found.
// Returns a slice of the absolute paths of all downloaded images, along with the number of missing images.
// 
// The number of missing images is the difference between the limit and the actual number of images downloaded. 
// This is only non-zero when the limit is higher than the number of downloadable images found.
func Download(query string, limit int, dir string, arguments ...string) (paths []string, missing int, err error) {
    dir, err = filepath.Abs(strings.ReplaceAll(dir, "\\", "/"))
    if err != nil {
        return []string{}, 0, err
    }

    urls, err := Urls(query, 0, arguments...)
    if err != nil {
        return []string{}, 0, err
    }

    var suffix int
    
    var i int
    for len(paths) < limit  {
        if i >= len(urls) {
            missing = limit-len(paths)
            break
        }

        url := urls[i]
        pat := path.Join(dir, query+strconv.Itoa(suffix)) + ".*"
        matches, _ := filepath.Glob(pat)
        for len(matches) > 0 {
            suffix++
            pat = path.Join(dir, query+strconv.Itoa(suffix)) + ".*"
            matches, _ = filepath.Glob(pat)
        }

        file, err := DownloadImage(url, dir, query+strconv.Itoa(suffix))
        for err != nil {
            i++
            if i >= len(urls) {
                missing = limit-len(paths)
                break
            }

            url = urls[i]
            file, err = DownloadImage(url, dir, query+strconv.Itoa(suffix))
        }

        paths = append(paths, file)
        i++
    }

    return paths, missing, nil
}

// Given the url of the image, the directory to download to, and the name of the file *without extension*, this will find the type of image and download it to the given directory.
// Warning: This will overwrite any image file with the same name, if the extension matches, so make sure to keep the name unique.
// You can check if a file with the name already exists with the following code:
//
//	import (
//	    "path"
//	    "path/filepath"
//	)
//
//	func exists(dir, name string) bool {
//	    pat := path.Join(dir, name) + ".*"
//	    matches, _ := filepath.Glob(pat)
//	    return len(matches) > 0
//	}
func DownloadImage(url, dir, name string) (imgpath string, err error) {
    dir, err = filepath.Abs(dir)
    if err != nil {
        return "", err
    }
    _, err = os.Stat(dir)
    if os.IsNotExist(err) {
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return "", err
        }
    }

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

// Checks if an error is an unpacking error. An unpacking error is generally thrown when Google changes their JSON structure, or on certain internet connections, when the specific header does not work.
// If you believe Google changed their JSON structure, please submit a bug report at https://github.com/commonkestrel/imagesearch/issues, and I will try to fix this asap.
func IsUnpackErr(err error) bool {
    return err == errUnpack
}

func buildUrl(query string, arguments []string) string {
    url := "https://www.google.com/search?tbm=isch&q=" + query

    if len(arguments) > 0 {
        url += "&tbs=ic:specific"
    }
    for _, argument := range arguments {
        url += "%2C" + argument
    }

    return url
}

func unpack(page string) ([]Image, error) {

    scriptStart := strings.LastIndex(page, "AF_initDataCallback")
    if scriptStart == -1 {
        return []Image{}, errUnpack
    }
    page = page[scriptStart:]

    startChar := strings.Index(page, "[")
    if startChar == -1 {
        return []Image{}, errUnpack
    }
    page = page[startChar:]

    endChar := strings.Index(page, "</script>") - 20
    if endChar == -1 {
        return []Image{}, errUnpack
    }
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
    // No idea why this works, but Google renders the page differently with this header. Credit to joeclinton1 on Github for this
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
