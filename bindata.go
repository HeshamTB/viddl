package main

import "embed"

//go:embed templates/*
var TemplatesFS embed.FS

//go:embed static/css/*
var PublicFS embed.FS

//go:embed assets/*
var AssetsFS embed.FS

//go:embed static/robots.txt
var RobotsFS embed.FS

