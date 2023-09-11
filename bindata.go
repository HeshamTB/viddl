package main

import "embed"

//go:embed templates/*
var TemplatesFS embed.FS

//go:embed static/css/* 
var PublicFS embed.FS
