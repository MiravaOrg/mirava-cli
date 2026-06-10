name: Release

on:
push:
tags:
- "v*"
workflow_dispatch:
inputs:
tag:
description: "Release tag (e.g. v1.0.0)"
required: true
type: string

permissions:
contents: write

env:
BINARY_NAME: mirava-cli

jobs:
build:
name: Build ${{ matrix.goos }}/${{ matrix.goarch }}
runs-on: ubuntu-latest
strategy:
fail-fast: false
matrix:
include:
- goos: linux
goarch: amd64
- goos: linux
goarch: arm64
- goos: darwin
goarch: amd64
- goos: darwin
goarch: arm64
- goos: windows
goarch: amd64
ext: .exe
- goos: windows
goarch: arm64
ext: .exe

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache-dependency-path: go.sum

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 0
        run: |
          ext="${{ matrix.ext }}"
          artifact="${{ env.BINARY_NAME }}-${{ matrix.goos }}-${{ matrix.goarch }}"
          go build -trimpath -ldflags="-s -w" -o "dist/${{ env.BINARY_NAME }}${ext}" .

          if [ "${{ matrix.goos }}" = "windows" ]; then
            cd dist
            zip "../${artifact}.zip" "${{ env.BINARY_NAME }}${ext}"
            cd ..
          else
            tar -czf "${artifact}.tar.gz" -C dist "${{ env.BINARY_NAME }}${ext}"
          fi

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: ${{ env.BINARY_NAME }}-${{ matrix.goos }}-${{ matrix.goarch }}
          path: |
            ${{ env.BINARY_NAME }}-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz
            ${{ env.BINARY_NAME }}-${{ matrix.goos }}-${{ matrix.goarch }}.zip
          if-no-files-found: error

release:
name: Publish release
needs: build
runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Determine tag
        id: version
        run: |
          if [ "${{ github.event_name }}" = "workflow_dispatch" ]; then
            echo "tag=${{ github.event.inputs.tag }}" >> "$GITHUB_OUTPUT"
          else
            echo "tag=${GITHUB_REF#refs/tags/}" >> "$GITHUB_OUTPUT"
          fi

      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          path: dist
          merge-multiple: true

      - name: Generate checksums
        run: |
          cd dist
          sha256sum * > checksums.txt

      - name: Create GitHub release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ steps.version.outputs.tag }}
          name: ${{ steps.version.outputs.tag }}
          generate_release_notes: true
          files: |
            dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}