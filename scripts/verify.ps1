$ErrorActionPreference = 'Stop'
$files = gofmt -l (Get-ChildItem -Recurse -Filter *.go | ForEach-Object FullName)
if ($files) { throw "gofmt required: $files" }
go test ./...
go test -race ./...
go vet ./...
Push-Location web
npm test -- --run
npm run typecheck
npm run build
Pop-Location
