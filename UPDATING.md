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
export NEW_VERSION=v0.0.12
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
