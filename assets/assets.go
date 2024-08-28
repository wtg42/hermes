package assets

import "embed"

// 打包用的 global variable
//
//go:embed fonts imgs/*
var StaticFiles embed.FS
