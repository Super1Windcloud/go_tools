default:
    @just --list

build-imgcat:
    go build -ldflags="-w -s" ./cmd/imgcat

build-clearlnk:
    go build -tags clearlnk -ldflags="-w -s" ./cmd/clearlnk

build-maven-search:
    go -C maven_search build -ldflags="-w -s" .

build:
    go build -ldflags="-w -s" ./cmd/imgcat
    go build -tags clearlnk -ldflags="-w -s" ./cmd/clearlnk
    go -C maven_search build -ldflags="-w -s" . 
    go -C exc build -ldflags="-w -s" . 


push:
    git add . && git commit -m "update" && git push repo main