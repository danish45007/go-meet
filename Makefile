osx:
	env GOOS=darwin GOARCH=arm go build
linux:
	env GOOS=linux GOARCH=arm go build
windows:
	env GOOS=windows GOARCH=arm go build