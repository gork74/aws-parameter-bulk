# Updating dependencies

```shell
go get -u ./...
```

```shell
go mod tidy
```

```shell
go test ./... -parallel=1
```

```shell
export NEW_VERSION=v0.0.15
```

```shell
sed -i '' "s/Version = \"v[0-9]*\.[0-9]*\.[0-9]*\"/Version = \"$NEW_VERSION\"/" conf/version.go
cat conf/version.go
```

```shell
git add .
git commit -m"feat: update dependencies"
git push
```

```shell
git tag "${NEW_VERSION?}"
git push --tags
```

In case the tag has to be deleted:
```shell
git tag -d "${NEW_VERSION?}"
git push origin ":refs/tags/${NEW_VERSION?}"
```

# Scan for vulnerabilities

Build
```shell
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/aws-parameter-bulk-darwin-arm64 main/main.go
```
Scan
```shell
trivy rootfs bin/aws-parameter-bulk-darwin-arm64
```
