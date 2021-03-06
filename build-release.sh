rm -rf ./release
mkdir ./release

GOOS=windows GOARCH=amd64 go build -o ./release/fitbitplot.exe ./
GOOS=linux GOARCH=amd64 go build -o ./release/fitbitplot-linux ./
GOOS=darwin GOARCH=amd64 go build -o ./release/fitbitplot-mac ./

cp README.md ./release/README.md

zip -r ./release/release.zip ./release