# imagesearch
[![Go Reference](https://pkg.go.dev/badge/github.com/jibble330/imagesearch.svg)](https://pkg.go.dev/github.com/jibble330/imagesearch)

A package designed to search Google Images based on the input query and arguments. These images may be protected under copyright, and you shouldn't do anything punishable with them, like using them for commercial use. 

---
# Arguments

There are 2 required parameters, along with a variety of different search arguments.
| Argument | Type | Description |
| --- | --- | --- |
| **query:** | *string* | The keyword(s) to search for.
| **limit** | *int* | The amount of images to search for. Cannot be greater then 100.  
| **arguments:** | *string* | Optional search parameters are passed through here.

## Search Arguments

These are used by passing ```imagesearch.{argument}.{option}``` in the ```arguments``` parameter. Only one option from each argument can be passed or Google will load the images without any arguments except the query.

| Argument | Options | Description |
| --- | --- | --- |
| **Color** | Red, Orange, Yellow, Green, Teal, Blue, Purple, Pink, White, Gray, Black, Brown | Filter images by the dominant color. |
| **ColorType** | Color, Grayscale, Transparent | Filter images by the color type, full color, grayscale, or transparent. |
| **License** | CreativeCommons, Other | Filter images by the usage license. |
| **Type** | Face, Photo, Clipart, Lineart, Animated | Filters by the type of images to search for. \**Not to be confused with search_format*\* |
| **Time** | PastDay, PastWeek, PastMonth, PastYear | Only finds images posted in the time specified. |
**AspectRatio** | Tall, Square, Wide, Panoramic | Specifies the aspect ratio of the images. |
**Format** | Jpg, Gif, Png, Bmp, Svg, Webp, Ico, Raw | Filters out images that are not a specified format. If you would like to download images as a specific format, use the download_format argument instead. |

---

## Credits
This library is inspired by the Python library [google-images-download](https://www.github.com/joeclinton1/google-images-download) created by **[hardikvasa](https://www.github.com/hardikvasa)** and maintained by **[joeclinton1](https://www.github.com/joeclinton1)**, but ported to **Go** and with some quality of life improvements, such as being able to retrieve urls as well. Essentially, this package is a port of the Python library [GoogleImageScraper](https://www.github.com/Jibble330/GoogleImageScraper) to **Go**.