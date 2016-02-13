# gin-cors

[![Build Status](https://travis-ci.org/tommy351/gin-cors.svg?branch=master)](https://travis-ci.org/tommy351/gin-cors)

CORS middleware for [Gin].

## Installation

``` bash
$ go get github.com/tommy351/gin-cors
```

## Usage

``` go
import (
    "github.com/gin-gonic/gin"
    "github.com/tommy351/gin-cors"
)

func main(){
    g := gin.New()
    g.Use(cors.Middleware(cors.Options{}))
}
```

[Gin]: http://gin-gonic.github.io/gin/
