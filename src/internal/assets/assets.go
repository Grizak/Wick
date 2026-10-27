package assets

import "embed"

//go:embed LICENSE
var License string

//go:embed changelog/*
var Changelog embed.FS
